package tests

type Service struct{
    name string
}

func Decl1() string {
    return "test"
}

func Decl2(decl1 string) {}

func Decl3(decl1 string) (decl3 *Service) {
    decl3 = &Service{name: decl1}
}
