package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis/v8"
	starterUsecases "github.com/mindaugasrmn/distributed-cowboys/core/starter/usecases"
	cowboyArray "github.com/mindaugasrmn/distributed-cowboys/helpers/cowboys"
	"github.com/mindaugasrmn/distributed-cowboys/helpers/kafkautils"
)

var (
	cowboyPath      = flag.String("file", "cowboys.json", "Path to cowboys file")
	cowboysInvolved = 0
)

func main() {
	flag.Parse()

	cowboys, err := cowboyArray.CowboysArray(*cowboyPath)
	if err != nil {
		panic(err)
	}
	cowboysInvolved = 0
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to initialize redis client: %s", err)

	}
	log.Println("connected to redis")

	mqp, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka-service:9092",
		"acks":              "all",
	})
	if err != nil {
		log.Fatalf("Kafka producer error: %s", err)

	}
	log.Println("created kafka producer")
	err = kafkautils.CreateTopic("kafka-service:9092", "attenders")
	if err != nil {
		log.Fatalf("Kafka CreateTopic failed, got error: %s", err)

	}
	log.Println("created topic attenders")
	mqc, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               "kafka-service:9092",
		"group.id":                        "attenders.GROUP",
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.partition.eof":            true,
		"auto.offset.reset":               "earliest"})

	if err != nil {
		log.Fatalf("Failed to create consumer: %s\n", err.Error())

	}
	err = mqc.Subscribe("attenders", nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %s\n", err.Error())
		os.Exit(1)
	}
	log.Println("created kafka consumer")

	starterUse := starterUsecases.NewStarterUsecases()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	t := time.Now()
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
				log.Printf("#### GOTDATA: %s", string(e.Value))
				for _, c := range cowboys {
					if strings.ToLower(c.Name) == strings.ToLower(string(e.Value)) {
						cowboysInvolved++
					}
				}
				if cowboysInvolved == len(cowboys) {
					podList, err := starterUse.GetPods(cowboysInvolved)
					if err != nil {
						fmt.Printf("Failed to get pod listd, got err: %s\n", err.Error())
					}

					if podList != nil && len(podList.Items) == cowboysInvolved {
						log.Println("got pods count:", len(podList.Items))
						t = time.Now().Add(20 * time.Second)
						for _, c := range podList.Items {
							value := fmt.Sprintf("%s:%d", c.Status.PodIP, c.Spec.Containers[0].Ports[0].ContainerPort)
							fmt.Println("adding:", strings.ToLower(c.Name), value)
							_, err = rdb.Set(context.Background(), strings.ToLower(c.Name), value, 0).Result()
							if err != nil {
								log.Println("rdb set got error:", err)
							}
							log.Println("publishing to", strings.ToLower(c.Spec.Containers[0].Name), " payload:", t)
							topic := strings.ToLower(c.Spec.Containers[0].Name)
							err = mqp.Produce(&kafka.Message{

								TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
								Value:          []byte(t.UTC().String()),
							}, nil)
							if err != nil {
								log.Fatalf("Failed to sent msg to topic attenders, got error: %s", err)
							}

						}
					}

				}

			case kafka.PartitionEOF:
				log.Printf("%% Reached %v\n", e)
			case kafka.Error:
				log.Printf("%% Error: %v\n", e)
			}
		}
	}

	mqc.Close()
	mqp.Close()
}
