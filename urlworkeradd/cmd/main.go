package main

import (
	"context"
	"datasplitter/external/timemanager"
	"envconfig"
	"log"
	"metrics"
	"strings"
	"urlworkeradd/external/adapters/storagerepo"
	"urlworkeradd/external/cacheadapters/redisrepo"
	"urlworkeradd/port/httpserver"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func getPortAddr(addr string) string {
	return ":" + strings.Split(addr, ":")[1]
}

// it'll be great to do smth with same code in workerget and workeradd
func main() {
	es, err := envconfig.NewEnvStorage()
	if err != nil {
		log.Fatal("urlWorkerAdder failed to connect to env storage")
	}
	dealerAddr, err := es.EnvGetVal("addr_dealer", "urlWorkerAddAddr")
	if err != nil {
		log.Fatal("urlWorkerAdder failed to get it's addr: ", err)
	}
	dataSplitterAddr, err := es.EnvGetVal("addr_dealer", "dataSplitterAddr")
	if err != nil {
		log.Fatal("urlWorkerAdder failed to get it's addr: ", err)
	}
	cRepoAddr, err := es.EnvGetVal("addr_dealer", "redisAddr")
	if err != nil {
		log.Fatal("urlWorkerAdder failed to get redis addr: ", err)
	}
	repoAddr, err := es.EnvGetVal("addr_dealer", "postgresAddr")
	if err != nil {
		log.Fatal("urlWorkerAdder failed to get postgres addr: ", err)
	}
	metricsAddr, err := es.EnvGetVal("addr_dealer", "urlWorkerAddMetricsAddr")
	if err != nil {
		log.Fatal("urlWorkerAdder failed to get metrics addr: ", err)
	}
	kafkaAddr, err := es.EnvGetVal("addr_dealer", "kafkaAddr")
	if err != nil {
		log.Fatal("urlWorkerAdder failed to get kafka broker addr: ", err)
	}
	kafkaTopic, err := es.EnvGetVal("kafka_topics_dealer", "urlAddQueueTopic")
	if err != nil {
		log.Fatal("urlWorkerAdder failed to get kafka topic: ", err)
	}

	parCtx := context.Background()
	eg, ctx := errgroup.WithContext(parCtx)
	lg, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to init logger")
	}
	cRepo := redisrepo.NewRepo(cRepoAddr, ctx, lg)
	//repo := postgresrepo.NewRepo(repoAddr, ctx, lg)
	repo := storagerepo.NewRepo(ctx, repoAddr, kafkaAddr, kafkaTopic, lg)

	metr, err := metrics.RunServer(ctx, eg, lg, getPortAddr(metricsAddr))
	if err != nil {
		log.Fatal("Metrics failed to run server: ", err)
	}
	err = httpserver.RunServer(ctx, lg, getPortAddr(dealerAddr), dataSplitterAddr, es, cRepo, repo, metr, timemanager.TmImpl{})
	if err != nil {
		lg.Error("Url worker adder failder server", zap.Error(err))
	}
	err = metrics.ShutDownServer(eg, lg)
	if err != nil {
		lg.Error("Metrics failder server", zap.Error(err))
	}
}
