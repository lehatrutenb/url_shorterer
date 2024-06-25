package consumer

import (
	"context"
	"strings"
	"urlworkeradd/external/adapters"
	"urlworkeradd/external/urls"

	"storage/internal/kafka"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func initConsumerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	return config
}

type MessageHandler struct {
	Ctx context.Context
	Db  adapters.Repository
	Eg  *errgroup.Group
	Lg  *zap.Logger
}

func NewMessageHandler(ctx context.Context, db adapters.Repository, eg *errgroup.Group, lg *zap.Logger) *MessageHandler {
	return &MessageHandler{ctx, db, eg, lg}
}

func (mh *MessageHandler) handleMessage(cms []*sarama.ConsumerMessage) []error {
	errs := make([]error, len(cms))
	for i, cm := range cms {
		u, err := urls.DecodeUrlFromBytes(cm.Value)
		errs[i] = err
		if err != nil {
			mh.Lg.Error("Failed to decode url", zap.Time("msg time stamp", cm.Timestamp))
			continue
		}
		if err := mh.Db.AddURL(u); err != nil {
			mh.Lg.Error("Failed to add msg to db", zap.Time("msg time stamp", cm.Timestamp))
			errs[i] = err
			continue
		}
		mh.Lg.Debug("Successfully added url to db", zap.String("short url", u.GetShortURL()), zap.String("long url", u.GetLongURL()), zap.Time("timestamp", cm.Timestamp))
	}
	return errs
}

func RunConsumer(ctx context.Context, lg *zap.Logger, msgHandler *MessageHandler, brokers string, group string, topics string) (sarama.ConsumerGroup, error) {
	cons := kafka.NewConsumer(msgHandler.handleMessage, lg)
	consGroup, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), group, initConsumerConfig())
	if err != nil {
		msgHandler.Lg.Error("Failed to init consumer group", zap.String("brokers", brokers), zap.String("group", group))
		return nil, err
	}

	msgHandler.Eg.Go(
		func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
					if err = consGroup.Consume(ctx, strings.Split(topics, ","), cons); err != nil {
						msgHandler.Lg.Error("Consumer cluster connection lost", zap.String("topics", topics))
						return err
					}
				}
			}
		},
	)
	return consGroup, nil
}
