package transport

type Transport interface {
	Listen() error
	Stop() error
	Send(addr string, data any) error
}
