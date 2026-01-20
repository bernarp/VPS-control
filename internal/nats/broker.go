package nats

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

var _ Broker = (*NatsBroker)(nil)

type NatsBroker struct {
	conn   *nats.Conn
	logger *zap.Logger
}

func NewNatsBroker(nc *NATSConn) *NatsBroker {
	return &NatsBroker{
		conn:   nc.Conn,
		logger: nc.logger.Named("broker"),
	}
}

func (b *NatsBroker) GetConn() *nats.Conn {
	return b.conn
}

func (b *NatsBroker) Publish(
	subject string,
	data any,
) error {
	payload := EventPayload[any]{
		ID:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		Version:   "1.0",
		Data:      data,
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	if err := b.conn.Publish(subject, bytes); err != nil {
		return fmt.Errorf("nats publish error: %w", err)
	}

	return nil
}

func Subscribe[T any](
	b Broker,
	subject string,
	handler func(payload EventPayload[T]) error,
	logger *zap.Logger,
) (*nats.Subscription, error) {
	return b.GetConn().Subscribe(
		subject, func(msg *nats.Msg) {
			var payload EventPayload[T]
			if err := json.Unmarshal(msg.Data, &payload); err != nil {
				logger.Error(
					"unmarshal failed",
					zap.String("subject", subject),
					zap.Error(err),
					zap.ByteString("raw_data", msg.Data),
				)
				return
			}
			if err := handler(payload); err != nil {
				logger.Error(
					"handler failed",
					zap.String("subject", subject),
					zap.String("event_id", payload.ID),
					zap.Error(err),
				)
			}
		},
	)
}

func QueueSubscribe[T any](
	b Broker,
	subject string,
	queue string,
	handler func(payload EventPayload[T]),
) (*nats.Subscription, error) {
	return b.GetConn().QueueSubscribe(
		subject, queue, func(msg *nats.Msg) {
			var payload EventPayload[T]
			if err := json.Unmarshal(msg.Data, &payload); err != nil {
				return
			}
			handler(payload)
		},
	)
}
