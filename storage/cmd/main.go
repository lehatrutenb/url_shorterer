package main

import (
	"context"
	"envconfig"
	"log"
	"os"
	"os/signal"
	"storage/internal/consumer"
	"syscall"
	"urlworkeradd/external/adapters/postgresrepo"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func connectToDbs(ctx context.Context, dbAddr string, lg *zap.Logger) *consumer.MessageHandler {
	eg, newCtx := errgroup.WithContext(ctx)
	db := postgresrepo.NewRepo(dbAddr, newCtx, lg)
	return consumer.NewMessageHandler(newCtx, db, eg, lg)
}

var group = "2"

func main() {
	es, err := envconfig.NewEnvStorage()
	if err != nil {
		log.Fatal("urlWorkerAdder failed to connect to env storage")
	}

	repoAddr, err := es.EnvGetVal("addr_dealer", "postgresAddr")
	if err != nil {
		log.Fatal("urlWorkerAdder failed to get postgres addr: ", err)
	}

	brokersAddr, err := es.EnvGetVal("addr_dealer", "kafkaAddr")
	if err != nil {
		log.Fatal("storage failed to get brokers addr: ", err)
	}

	topics, err := es.EnvGetVal("kafka_topics_dealer", "urlAddQueueTopic")
	if err != nil {
		log.Fatal("storage failed to get topic: ", err)
	}

	lg, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to init logger")
	}
	lg.Info("Starting storage", zap.String("brokers", brokersAddr), zap.String("topics", topics))
	sarama.Logger = zap.NewStdLog(lg.With(zap.String("storage", "sarama")))

	ctx, cncl := context.WithCancel(context.Background())
	msgHandler := connectToDbs(ctx, repoAddr, lg)
	consumer, err := consumer.RunConsumer(ctx, lg, msgHandler, brokersAddr, group, topics)
	if err != nil {
		log.Fatal(err)
	}

	lg.Info("Sarama is running")

	sigterm := make(chan os.Signal, 2)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	keepRunning := true
	for keepRunning {
		select {
		case <-ctx.Done():
			msgHandler.Lg.Info("Ctx cancelled")
			keepRunning = false
		case <-sigterm:
			msgHandler.Lg.Info("Get sigterm")
			keepRunning = false
			cncl()
		}
	}

	if err = consumer.Close(); err != nil {
		msgHandler.Lg.Error("Failed to shut down consumer", zap.Error(err))
		return
	}

	if err := msgHandler.Eg.Wait(); err != nil {
		msgHandler.Lg.Error("Storage shut down with error", zap.Error(err))
		return
	}

	if err = msgHandler.Db.CloseRepo(); err != nil {
		msgHandler.Lg.Error("Failed to shut down db", zap.Error(err))
		return
	}

	msgHandler.Lg.Info("Storage shut down successfully")
}
