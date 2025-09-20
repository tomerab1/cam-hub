package events

import "context"

type Handler = func(ctx context.Context, m Message) AckAction
type AckAction = int

const (
	Ack AckAction = iota
	NackRequeue
	NackDiscard
)

type BusIface interface {
	Publish(ctx context.Context, exc, key string, body []byte, headrs map[string]any) error
	Consume(ctx context.Context, queue string, h Handler) error
	Close() error
}

type Message struct {
	Body        []byte
	Headers     map[string]any
	Key         string
	Redelivered bool
}
