package nats

import (
	"VPS-control/internal/config"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NATSConn struct {
	Conn   *nats.Conn
	logger *zap.Logger
}

func NewConnection(
	cfg config.NATSConfig,
	logger *zap.Logger,
) (*NATSConn, error) {
	log := logger.Named("nats")

	opts := []nats.Option{
		nats.Name("VPS_CONTROL_API"),
		nats.Timeout(cfg.Timeout),
		nats.MaxReconnects(-1),
		nats.DisconnectErrHandler(
			func(
				nc *nats.Conn,
				err error,
			) {
				log.Warn("NATS disconnected", zap.Error(err))
			},
		),
		nats.ReconnectHandler(
			func(nc *nats.Conn) {
				log.Info("NATS reconnected", zap.String("url", nc.ConnectedUrl()))
			},
		),
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats connection failed: %w", err)
	}

	log.Info("Connected to NATS", zap.String("url", cfg.URL))

	return &NATSConn{
		Conn:   nc,
		logger: log,
	}, nil
}

func (n *NATSConn) Close() {
	if n.Conn != nil {
		n.logger.Info("Closing NATS connection")
		n.Conn.Close()
	}
}
