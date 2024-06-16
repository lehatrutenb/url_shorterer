package main

import (
	"context"
	timemanager "datasplitter/internal/time_manager"
	"datasplitter/port/httpserver"
	"envconfig"
	"log"
	"strconv"
	"strings"
)

func main() {
	es, err := envconfig.NewEnvClientStorage()
	if err != nil {
		log.Fatal("datasplitter failed to connect to env storage")
	}
	bSz, err := es.EnvGetVal("saved_vals", "short_urls_batch_size")
	if err != nil {
		log.Fatal("datasplitter failed to get batch size")
	}
	urlBatchSize, err := strconv.Atoi(bSz)
	if err != nil {
		log.Fatal("datasplitter invalid batch size: ", err)
	}

	dealerAddr, err := es.EnvGetVal("addr_dealer", "dataSplitterAddr")
	if err != nil {
		log.Fatal("datasplitter failed to get splitter addr: ", err)
	}
	httpserver.RunServer(context.Background(), ":"+strings.Split(dealerAddr, ":")[1], urlBatchSize, timemanager.TmImpl{}, es)
}
