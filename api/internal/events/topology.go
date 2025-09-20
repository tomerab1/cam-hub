package events

type DeclarerIface interface {
	DeclareExchange(name, kind string, durable bool) error
	DeclareQueue(name string, durable bool, args map[string]any) error
	Bind(queue, exch, key string, args map[string]any) error
}
