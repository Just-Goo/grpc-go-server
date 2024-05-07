package application

// application layer contains the business logic and the implementation of the interfaces in the port layer

type HelloService struct {
}

func (h *HelloService) GenerateHello(name string) string {
	return "Hello " + name
}
