package gonuts

//General Type of any component
type ComponentType interface{}

type ArgumentType struct {
    Types []ComponentType
}

type OutputType struct {
    Types []ComponentType
}

type FunctionType struct {
    Args *ArgumentType
    Output *OutputType
}


type PointerType struct {
    PointerTo ComponentType
}

type BuiltinType struct {
    Name string
}

type NamedType struct {
    Name string
    PackagePath string
}

type ArrayType struct {
    Type ComponentType
}

type InterfaceMethod struct {
    Name string
    Args *ArgumentType
    Output *OutputType
}

type InterfaceType struct {
    Methods []*InterfaceMethod
}

type Component struct {
    Name string
    Type ComponentType
    Factory *ComponentFactory
}

type ComponentFactory struct {
    FactoryName ComponentType
    Dependencies []string
    Produces []*Component
    ProducesError bool
}

