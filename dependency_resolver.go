package gonuts

import "fmt"

type ResolvedDependencyGraph struct {
    Declarations []*ComponentDeclarations
}

type ComponentDeclarations struct {
    Components []*Component
    Declaration *FactoryInvocation
}

type FactoryInvocation struct {
    Factory *ComponentFactory
    ComponentArgs []*Component
}

type DependencyGraphResolver struct {
    registry *ComponentRegistry
}

//#Factory
func NewDependencyGraphResolver(registry *ComponentRegistry) (dependency_resolver *DependencyGraphResolver) {
    return &DependencyGraphResolver{registry: registry}
}

func (self *DependencyGraphResolver) ResolveDependencies() (*ResolvedDependencyGraph, error) {
    components := self.registry.Components()
    declarations := make([]*ComponentDeclarations, 0, len(components))
    declaredComponents := make(map[string]bool)
    for {
        declaredComponentCnt := len(declaredComponents)
        if declaredComponentCnt >= len(components) {
            //done
            break
        }
        for _, component := range components {
            if !declaredComponents[component.Name] {
                if component.Factory != nil {
                    missing_dependency := false
                    factory := component.Factory
                    componentArgs := make([]*Component, len(factory.Dependencies))
                    for i, dependency := range factory.Dependencies {
                        if !declaredComponents[dependency] {
                            missing_dependency = true
                            break
                        }
                        componentArgs[i] = self.registry.Component(dependency)
                    }
                    if !missing_dependency {
                        componentDeclarations := &ComponentDeclarations{Components: make([]*Component,len(factory.Produces))}
                        for i, producedComponent := range factory.Produces {
                            declaredComponents[producedComponent.Name] = true
                            componentDeclarations.Components[i] = producedComponent
                        }
                        componentDeclarations.Declaration = &FactoryInvocation{Factory:factory, ComponentArgs: componentArgs}
                        declarations = append(declarations, componentDeclarations)
                    }
                }
            }
        }
        if len(declaredComponents) <= declaredComponentCnt {
            //find the dependencies that we couldn't declare
            missing := make([]string, 0, len(components) - len(declaredComponents))
            for _, component := range components {
                if !declaredComponents[component.Name] {
                    missing = append(missing,component.Name)
                }
            }
            //ERROR
            return nil, fmt.Errorf("Unresolvable dependensice, likely due to circular or undeclared dependencies. %s", missing)
        }
    }
    return &ResolvedDependencyGraph{Declarations:declarations}, nil
}
