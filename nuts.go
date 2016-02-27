package main

import (
	"go/parser"
    "go/ast"
    "go/token"
    "fmt"
)

func main() {
	//by default process the current directory
	package_map,err := parser.ParseDir(token.NewFileSet(),"/Users/chris/Code/Go/src/github.com/routinesub/go-nuts/tests",
        nil,parser.ParseComments)
	if err != nil{
        fmt.Printf("Error %s", err.Error())
        return;
	}
    components := make([]component,0,5)
    //discovery
	for pkg_name, package_ast := range package_map {
        fmt.Printf("import %s", pkg_name)
        for _,file := range package_ast.Files {
            if ast.FileExports(file) {
                for _, decl := range file.Decls {
                    func_decl, ok := decl.(*ast.FuncDecl)
                    if ok && func_decl.Recv == nil {
                        comp := &component{name:func_decl.Name.Name}
                        fmt.Println(func_decl.Name.Name)
                        dependencies := make([]string,len(func_decl.Type.Params.List))
                        for i,field := range func_decl.Type.Params.List {
                            dependencies[i] = field.Names[0].Name
                        }
                        comp.dependencies = dependencies
                        components = append(components,comp)
                    }
                }
            }
        }
	}
    //output
    declared := make(map[string]bool)
    for {
        cur_decl_cnt := len(declared)
        if len(components) == cur_decl_cnt {
            break
        }
        for _, comp := range components {
            if !declared[comp.name] {
                missing_dependency := false
                for _, dependency := range comp.dependencies {
                    if !declared[dependency] {
                        missing_dependency = true
                        break
                    }
                }
                if !missing_dependency {
                    fmt.Print(comp.name)
                    fmt.Print(" := ")
                    fmt.Print(comp.name)
                    fmt.Print("(")
                    for i,dependency := range comp.dependencies {
                        fmt.Print(dependency)
                        if i < len(comp.dependencies) - 1 {
                            fmt.Print(", ")
                        }
                    }
                    fmt.Print(")")
                    fmt.Println()
                }
            }
        }
        if len(declared) == cur_decl_cnt {
            fmt.Println("Error! unmet dependencies")
            break
        }
    }
}

type component struct {
    name string
    dependencies []string
}