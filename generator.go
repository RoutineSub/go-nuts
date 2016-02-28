package gonuts

import (
    "os"
    "text/template"
    "fmt"
    "strings"
)

const app_template =
`package {{.Package}}

{{range .Imports}}
import {{if .Alias}}{{.Alias}} {{end}}"{{.Path}}"
{{end}}

type app struct {
    {{range .Components}}
    {{.Name}} {{.Type}}
    {{end}}
}

func create_app() (a *app, err error) {
    a = &app{}
    {{range .Declarations}}
    {{join .Components ","}}{{if .HasError}}, err{{end}} = {{.Function}}({{join .Args ","}})
    {{if .HasError}}if err != nil {
        return nil, err
    }{{end}}
    {{end}}
    return
}
`

type source_file struct {
    Package string
    Imports []*import_stmnt
    Components []*component_member
    Declarations []*declaration
}

type import_stmnt struct {
    Alias string
    Path string
}

type component_member struct {
    Name string
    Type string
}

type declaration struct {
    Components []string
    HasError bool
    Function string
    Args []string
}

type AppGenerator struct {
    template *template.Template
    graphResolver *DependencyGraphResolver
    registry *ComponentRegistry
}

//#Factory
func NewAppGenerator(dependency_resolver *DependencyGraphResolver, registry *ComponentRegistry) (app_generator *AppGenerator, err error) {
    if temp, err := template.New("app").Funcs(template.FuncMap{"join": strings.Join}).Parse(app_template); err != nil {
        return nil, err
    } else {
        app_generator = &AppGenerator{template: temp, graphResolver:dependency_resolver, registry:registry}
    }
    return
}

func (self *AppGenerator) GenerateApp() error {
    resolvedGraph, err := self.graphResolver.ResolveDependencies()
    if err != nil {
         return err
    }

    components := self.registry.Components()

    source := &source_file{Package:"main",
        Components:make([]*component_member,len(components)),
        Declarations:make([]*declaration,len(resolvedGraph.Declarations)),
        Imports: make([]*import_stmnt,0)}

    import_map := make(map[string]string)
    defined_imports := make(map[string]struct{})
    get_or_create_import := func (pkg_path string) string {
        if import_map[pkg_path] != "" {
            return import_map[pkg_path]
        } else {
            default_alias := get_default_import_alias(pkg_path)
            alias := default_alias
            _, defined := defined_imports[alias]
            for i := 0; defined; i++ {
                alias = fmt.Sprintf("%s%i",default_alias,i)
                _, defined = defined_imports[alias]
            }
            import_map[pkg_path] = alias
            defined_imports[alias] = struct{}{}
            stmnt := &import_stmnt{Path:pkg_path}
            if alias != default_alias {
                stmnt.Alias = alias
            }
            source.Imports = append(source.Imports,stmnt)
            return alias
        }
    }

    for i,component := range components {
        source.Components[i] = &component_member{Name: component.Name,
            Type: generate_component_type_string(component.Type, get_or_create_import)}
    }

    for i, componentDecls := range resolvedGraph.Declarations {
        source.Declarations[i] = &declaration{Components:make([]string,len(componentDecls.Components)),
            Args:make([]string,len(componentDecls.Declaration.ComponentArgs)),
            Function: generate_component_type_string(componentDecls.Declaration.Factory.FactoryName, get_or_create_import),
            HasError:componentDecls.Declaration.Factory.ProducesError }
        for j,comp := range componentDecls.Components {
            source.Declarations[i].Components[j] = fmt.Sprintf("a.%s", comp.Name)
        }
        for j,arg := range componentDecls.Declaration.ComponentArgs {
            source.Declarations[i].Args[j] = fmt.Sprintf("a.%s", arg.Name)
        }
    }

    return self.template.Execute(os.Stdout,source)
}

func generate_component_type_string(t ComponentType, get_import func(string) string) string{
    switch t := t.(type) {
    case *BuiltinType:
        return t.Name
    case *ArrayType:
        return fmt.Sprintf("[]%s",generate_component_type_string(t.Type, get_import))
    case *PointerType:
        return fmt.Sprintf("*%s",generate_component_type_string(t.PointerTo, get_import))
    case *NamedType:
        return fmt.Sprintf("%s.%s",get_import(t.PackagePath), t.Name)
    case *ArgumentType:
        args := make([]string,len(t.Types))
        for i,arg_t := range t.Types {
            args[i] = generate_component_type_string(arg_t, get_import)
        }
        return strings.Join(args,"")
    case *OutputType:
        if len(t.Types) == 0 {
            return ""
        } else if len(t.Types) == 1 {
            return fmt.Sprintf(" %s", generate_component_type_string(t.Types[0], get_import))
        } else {
            outputs := make([]string, len(t.Types))
            for i,out_t := range t.Types {
                outputs[i] = generate_component_type_string(out_t, get_import)
            }
            return fmt.Sprintf(" (%s)", strings.Join(outputs,","))
        }
    case *FunctionType:
        return fmt.Sprintf("func(%s)%s",generate_component_type_string(t.Args,get_import),
            generate_component_type_string(t.Output,get_import))
    case *InterfaceType:
        methods := make([]string, len(t.Methods))
        for i, method := range t.Methods {
            methods[i] = fmt.Sprintf("%s(%s)%s", method.Name,
                generate_component_type_string(method.Args,get_import),
                generate_component_type_string(method.Output,get_import))
        }
        return fmt.Sprintf("interface { %s }",strings.Join(methods,"; "))
    }
    return ""
}

func get_default_import_alias(pkg_path string) string {
    i := strings.LastIndex(pkg_path,"/")
    return strings.Replace(pkg_path[i+1:],"-","",1)
}
