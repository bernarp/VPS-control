package nats

import (
	"time"

	"github.com/nats-io/nats.go"
)

type Connection interface {
	Close()
}

type Broker interface {
	Publish(
		subject string,
		data any,
	) error
	GetConn() *nats.Conn
}

type EventPayload[T any] struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Data      T         `json:"data"`
}
