package gonuts

import (
    "go/ast"
    "strings"
    "fmt"
)

var builtinTypes = map[string]struct{}{"uint8":{},"uint16":{},"uint32":{},"uint64":{},"int8":{},"int16":{},"int32":{},
    "int64":{},"float32":{},"float64":{},"complex64":{},"complex128":{},"byte":{},"rune":{},"unit":{},"int":{},
    "unitptr":{},"bool":{},"string":{}}

type Registry interface {
    RegisterComponent(*Component)
}

type ComponentDiscovery struct {
    registry Registry
}

//#Factory
func NewComponentDiscovery(registry Registry) (componentDiscovery *ComponentDiscovery){
    return &ComponentDiscovery{registry: registry}
}

func (self *ComponentDiscovery) DiscoverComponents(file *ast.File, current_path string) {
    //create a map of all the imports
    import_map := parse_import_map(file)

    //look at the the declarations in a file
    for _, decl := range file.Decls {
        switch decl := decl.(type){
        case *ast.FuncDecl:
            parse_factory_function(decl, self.registry, import_map, current_path)
        }

    }
}

func parse_import_map (file *ast.File) map[string]string {
    impt_map := make(map[string]string)
    for _,impt := range file.Imports {
        path := strings.Trim(impt.Path.Value, "\"")
        if impt.Name != nil {
            impt_map[impt.Name.Name] = path
        } else {
            pos := strings.LastIndex(path, "/")
            name := path[pos+1:]
            impt_map[name] = path
        }
    }
    return impt_map
}

func parse_factory_function(func_decl *ast.FuncDecl, registry Registry, import_map map[string]string, current_path string) {
    if is_function_factory(func_decl) {
        dependencies := make([]string, len(func_decl.Type.Params.List))
        for i, dependency := range func_decl.Type.Params.List {
            dependencies[i] = dependency.Names[0].Name
        }

        results := func_decl.Type.Results.List
        is_error := is_type_error(results[len(results) - 1].Type)
        if is_error {
            results = results[:len(results) - 1]
        }

        components := make([]*Component,len(results))
        for i, output := range results {
            var name string
            if len(output.Names) < 1 {
                if len(results) == 1 {
                    name = func_decl.Name.Name
                } else {
                    name = fmt.Sprintf("%s%i",func_decl.Name.Name, i)
                }
            } else {
                name = output.Names[0].Name
            }
            component_type := determine_component_type(output.Type, import_map, current_path)
            component := &Component{Name: name, Type:component_type}
            components[i] = component
        }
        componentFactory := &ComponentFactory{
            FactoryName:&NamedType{Name:func_decl.Name.Name, PackagePath:current_path}, Dependencies: dependencies,
            Produces:components, ProducesError: is_error}
        for _, component := range components {
            component.Factory = componentFactory
            registry.RegisterComponent(component)
        }
    }
}

func is_function_factory(func_decl *ast.FuncDecl) bool {
    //look for factory comments, no receiver, and at least one result
    if func_decl.Doc != nil && func_decl.Recv == nil && len(func_decl.Type.Results.List) > 0 {
        for _, comment := range func_decl.Doc.List {
            if strings.Contains(comment.Text, "#Factory") {
                return true
            }
        }
    }
    return false
}

func is_type_error(t_expr ast.Expr) bool {
    ident, ok := t_expr.(*ast.Ident)
    return ok && ident.Name == "error"
}

func determine_component_type(t_expr ast.Expr, import_map map[string]string, current_package string) ComponentType {

    switch t_expr := t_expr.(type) {

    case *ast.Ident:
        if _,ok := builtinTypes[t_expr.Name]; ok {
            return &BuiltinType{Name:t_expr.Name}
        } else {
            return &NamedType{Name: t_expr.Name, PackagePath:current_package}
        }

    case *ast.StarExpr:
        return &PointerType{PointerTo:determine_component_type(t_expr.X, import_map, current_package)}

    case *ast.SelectorExpr:
        if selector, ok := t_expr.X.(*ast.Ident); ok{
            return &NamedType{Name: t_expr.Sel.Name, PackagePath:import_map[selector.Name]}
        } else {
            return nil
        }

    case *ast.FuncType:

        var args []ComponentType
        if t_expr.Params != nil {
            args = make([]ComponentType, len(t_expr.Params.List))
            for i, f := range t_expr.Params.List {
                args[i] = determine_component_type(f.Type, import_map, current_package)
            }
        } else {
            args = make([]ComponentType,0)
        }

        var out []ComponentType
        if t_expr.Results != nil {
            out = make([]ComponentType, len(t_expr.Results.List))
            for i, f := range t_expr.Results.List {
                out[i] = determine_component_type(f.Type, import_map, current_package)
            }
        } else {
            out = make([]ComponentType, 0)
        }

        return &FunctionType{Args:&ArgumentType{Types:args}, Output:&OutputType{Types:out}}

    case *ast.ArrayType:
        return &ArrayType{Type:determine_component_type(t_expr.Elt, import_map, current_package)}

    case *ast.InterfaceType:
        methods := make([]*InterfaceMethod,len(t_expr.Methods.List))
        for i, method := range t_expr.Methods.List{
            m_type,ok := method.Type.(*ast.FuncType)
            if ok {
                inputs := make([]ComponentType, len(m_type.Params.List))
                for j, param := range m_type.Params.List {
                    inputs[j] = determine_component_type(param.Type, import_map, current_package)
                }
                outputs := make([]ComponentType, len(m_type.Results.List))
                for j,result := range m_type.Results.List {
                    outputs[j] = determine_component_type(result.Type, import_map, current_package)
                }
                methods[i] = &InterfaceMethod{Name:method.Names[0].Name,Args:&ArgumentType{Types:inputs},
                    Output:&OutputType{Types:outputs}}
            }
        }
        return &InterfaceType{Methods:methods}
    }
    return nil
}
