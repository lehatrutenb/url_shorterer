package main

import (
	"context"
	"datasplitter/external/timemanager"
	"envconfig"
	"log"
	"strings"
	"urlworkeradd/external/adapters/postgresrepo"
	"urlworkeradd/external/cacheadapters/redisrepo"
	"urlworkerget/port/httpserver"

	"go.uber.org/zap"
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
	dealerAddr, err := es.EnvGetVal("addr_dealer", "urlWorkerGetAddr")
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

	ctx := context.Background()
	lg, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to init logger")
	}
	cRepo := redisrepo.NewRepo(cRepoAddr, ctx, lg)
	repo := postgresrepo.NewRepo(repoAddr, ctx, lg)

	httpserver.RunServer(context.Background(), lg, ":"+strings.Split(dealerAddr, ":")[1], es, cRepo, repo, timemanager.TmImpl{}, 100)
}
