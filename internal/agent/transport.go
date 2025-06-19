package agent

type Transport interface {
	ServeHTTP()
}
