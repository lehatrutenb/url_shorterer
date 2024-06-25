package main

import (
	"envconfig"
	"log"
)

func main() {
	es, err := envconfig.NewEnvStorage()
	if err != nil {
		log.Fatal("failed to connect to env storage: ", err)
	}
	errs := make([]error, 9)
	errs[0] = es.EnvUpdateVal("addr_dealer", "postgresAddr", "postgres://postgres:postgres@postgres:5432/postgres")
	errs[1] = es.EnvUpdateVal("addr_dealer", "dataSplitterAddr", "datasplitter:9092")
	errs[1] = es.EnvUpdateVal("addr_dealer", "urlWorkerAddAddr", "urlworkeradd:9090")
	errs[1] = es.EnvUpdateVal("addr_dealer", "urlWorkerGetAddr", "urlworkerget:9091")
	errs[2] = es.EnvUpdateVal("addr_dealer", "kafkaAddr", "kafka1:9093")
	errs[3] = es.EnvUpdateVal("addr_dealer", "redisAddr", "redis:6379")
	errs[4] = es.EnvUpdateVal("addr_dealer", "urlWorkerAddMetricsAddr", "urlworkeradd:9094")
	errs[5] = es.EnvUpdateVal("addr_dealer", "urlWorkerGetMetricsAddr", "urlworkerget:9095")
	errs[6] = es.EnvUpdateVal("kafka_topics_dealer", "urlAddQueueTopic", "links.add")
	errs[7] = es.EnvUpdateVal("saved_vals", "last_batch_index", "0")
	errs[8] = es.EnvUpdateVal("saved_vals", "short_urls_batch_size", "1000000")

	for _, err := range errs {
		if err != nil {
			log.Fatal("failed to init base envs: ", err)
		}
	}
}
