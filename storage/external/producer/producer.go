package producer

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type Producer struct {
	producer sarama.AsyncProducer
	topic    string
}

func NewProducer(ctx context.Context, kafAddr string, lg *zap.Logger, topic string) (*Producer, error) {
	lg = lg.With(zap.String("app", "producer"))
	sarama.Logger = zap.NewStdLog(lg)

	producer, err := createProducer([]string{kafAddr})
	if err != nil {
		lg.Error("Failed to create kafka producer", zap.Error(err), zap.String("kafka broker addr", kafAddr), zap.String("kafka topic", topic))
		return nil, err
	}
	go func() {
		for {
			select {
			case err := <-producer.Errors():
				lg.Warn("Kafka producer error", zap.Error(err.Err))
			case <-ctx.Done():
				if err := producer.Close(); err != nil {
					lg.Warn("Kafka producer close error", zap.Error(err))
				}
			}
		}
	}()
	return &Producer{producer: producer, topic: topic}, nil
}

func (pr *Producer) WriteMessage(m []byte) {
	pr.producer.Input() <- &sarama.ProducerMessage{
		Topic:     pr.topic,
		Value:     sarama.ByteEncoder(m),
		Timestamp: time.Now(),
	}
}

func createProducer(brokerList []string) (sarama.AsyncProducer, error) {
	c := sarama.NewConfig()
	c.Version = sarama.DefaultVersion
	c.Producer.RequiredAcks = sarama.WaitForLocal
	c.Producer.Compression = sarama.CompressionSnappy
	c.Producer.Flush.Frequency = 500 * time.Millisecond

	producer, err := sarama.NewAsyncProducer(brokerList, c)
	if err != nil {
		return nil, err
	}

	return producer, nil
}
