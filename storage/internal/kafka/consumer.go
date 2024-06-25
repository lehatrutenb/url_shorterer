package kafka

import (
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type MessageHandlerFunc func(message []*sarama.ConsumerMessage) []error

type Consumer struct {
	handler      MessageHandlerFunc
	batch        []*sarama.ConsumerMessage
	bMaxInterval time.Duration
	bMaxSize     int
	lg           *zap.Logger
}

func NewConsumer(handler MessageHandlerFunc, lg *zap.Logger) *Consumer {
	return &Consumer{
		handler:      handler,
		batch:        make([]*sarama.ConsumerMessage, 0),
		bMaxInterval: time.Second * 100,
		bMaxSize:     100,
		lg:           lg.With(zap.String("app", "storage")),
	}
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) handleBatch(session sarama.ConsumerGroupSession) {
	errs := c.handler(c.batch)
	for i, msg := range c.batch {
		if errs[i] == nil {
			session.MarkMessage(msg, "")
			continue
		}
		c.lg.Error("failed to handle message: ", zap.Error(errs[i]))
	}
	c.batch = make([]*sarama.ConsumerMessage, 0)
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ticker := time.NewTicker(c.bMaxInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.handleBatch(session)
		case msg, ok := <-claim.Messages():
			if !ok {
				c.lg.Warn("message channel closed")
				return nil
			}
			c.batch = append(c.batch, msg)
			if len(c.batch) == c.bMaxSize {
				c.handleBatch(session)
			}
		case <-session.Context().Done():
			session.Commit()
			return nil
		}
	}
}
