package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	logRepository "github.com/mindaugasrmn/distributed-cowboys/core/logs/repository"
	logUsecases "github.com/mindaugasrmn/distributed-cowboys/core/logs/usecases"
	"github.com/mindaugasrmn/distributed-cowboys/helpers/kafkautils"
)

// Grpc server for logs starts here
func main() {
	flag.Parse()

	storeFile, err := os.OpenFile(
		path.Join(".", fmt.Sprintf("%s%s", "log", ".txt")),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return
	}
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	err = kafkautils.CreateTopic("kafka-service:9092", "log")
	if err != nil {
		log.Fatalf("Kafka CreateTopic failed, got error: %s", err)

	}
	mqc, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               "kafka-service:9092",
		"group.id":                        "log.GROUP",
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.partition.eof":            true,
		"auto.offset.reset":               "earliest"})

	if err != nil {
		log.Fatalf("Failed to create consumer: %s\n", err.Error())
		os.Exit(1)
	}
	err = mqc.Subscribe("log", nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %s\n", err.Error())
		os.Exit(1)
	}
	store := logRepository.NewStore(storeFile)
	logUse := logUsecases.NewLog(store)
	run := true
	for run {
		select {
		case sig := <-sigchan:
			log.Printf("Caught signal %v: terminating\n", sig)
			run = false

		case ev := <-mqc.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				log.Printf("%% %v\n", e)

				mqc.Assign(e.Partitions)
			case kafka.RevokedPartitions:
				log.Printf("%% %v\n", e)

				mqc.Unassign()
			case *kafka.Message:
				log.Println(string(e.Value))

				_, err := logUse.Append(string(e.Value))
				if err != nil {
					log.Printf("Failed to append to log, got error: %s\n", err.Error())
				}

			case kafka.PartitionEOF:
				//log.Printf("%% Reached %v\n", e)
			case kafka.Error:
				log.Printf("%% Error: %v\n", e)
			}
		}
	}

	log.Println("Closing consumer and producer")
	mqc.Close()
}
