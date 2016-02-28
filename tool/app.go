package main


import "github.com/routinesub/go-nuts"


type app struct {

    componentDiscovery *gonuts.ComponentDiscovery

    registry *gonuts.ComponentRegistry

    dependency_resolver *gonuts.DependencyGraphResolver

    app_generator *gonuts.AppGenerator

}

func create_app() (a *app, err error) {
    a = &app{}

    a.registry = gonuts.NewComponentRegistry()


    a.dependency_resolver = gonuts.NewDependencyGraphResolver(a.registry)


    a.app_generator, err = gonuts.NewAppGenerator(a.dependency_resolver,a.registry)
    if err != nil {
        return nil, err
    }

    a.componentDiscovery = gonuts.NewComponentDiscovery(a.registry)


    return
}