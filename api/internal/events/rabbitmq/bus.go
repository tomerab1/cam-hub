package rabbitmq

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	"tomerab.com/cam-hub/internal/events"
)

type AMQPBus struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
}

func NewBus(url string) (*AMQPBus, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &AMQPBus{conn: conn, ch: ch}, nil
}

func (bus *AMQPBus) Publish(ctx context.Context, exch, key string, body []byte, headrs map[string]any) error {
	return nil
}

func (bus *AMQPBus) Consume(ctx context.Context, queue string, h events.Handler) error {
	msgs, err := bus.ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for m := range msgs {
			msg := events.Message{
				Body:        m.Body,
				Key:         m.RoutingKey,
				Headers:     m.Headers,
				Redelivered: m.Redelivered,
			}

			switch h(ctx, msg) {
			case events.Ack:
				_ = m.Ack(false)
			case events.NackRequeue:
				_ = m.Nack(false, true)
			case events.NackDiscard:
				_ = m.Nack(false, false)
			}
		}
	}()

	return nil
}

func (bus *AMQPBus) Close() error {
	if err := bus.ch.Close(); err != nil {
		_ = bus.conn.Close()
		return err
	}
	return bus.conn.Close()
}
