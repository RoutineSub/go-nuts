package gonuts

type ComponentRegistry struct {
    components map[string]*Component
}

//#Factory
func NewComponentRegistry() (registry *ComponentRegistry) {
    registry = &ComponentRegistry{components:make(map[string]*Component)}
    return
}

func (self *ComponentRegistry) RegisterComponent(component *Component) {
    self.components[component.Name] = component
}

func (self *ComponentRegistry) Component(name string) *Component {
    return self.components[name]
}

func (self *ComponentRegistry) Components() []*Component {
    components := make([]*Component,len(self.components))
    i := 0
    for _,component := range self.components {
        components[i] = component
        i++
    }
    return components
}
