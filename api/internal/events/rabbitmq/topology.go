package rabbitmq

func (bus *AMQPBus) DeclareExchange(name, kind string, durable bool) error {
	return bus.ch.ExchangeDeclare(name, kind, durable, false, false, false, nil)
}

func (bus *AMQPBus) DeclareQueue(name string, durable bool, args map[string]any) error {
	_, err := bus.ch.QueueDeclare(name, durable, false, false, false, args)
	return err
}

func (bus *AMQPBus) Bind(queue, exch, key string, args map[string]any) error {
	return bus.ch.QueueBind(queue, key, exch, false, args)
}
