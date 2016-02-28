package main

import (
	"go/parser"
    "go/token"
    "fmt"
)

func main() {
	//by default process the current directory
	pkgs,err := parser.ParseDir(token.NewFileSet(),"/Users/chris/Code/Go/src/github.com/routinesub/go-nuts",
        nil, parser.ParseComments)
	if err != nil{
        fmt.Printf("Error %s", err.Error())
        return;
	}
    if app, err := create_app(); err != nil {
        fmt.Println(err)
    } else {
        for _, pkg := range pkgs {
            for _, file := range pkg.Files {
                app.componentDiscovery.DiscoverComponents(file, "github.com/routinesub/go-nuts")
            }
        }
        err := app.app_generator.GenerateApp()
        if err != nil {
            fmt.Println(err)
        }
    }
}
