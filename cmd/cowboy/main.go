package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"

	//cowbloyDeliverygRPC "github.com/mindaugasrmn/distributed-cowboys/core/cowboy/delivery/grpc"
	cowboyUsecases "github.com/mindaugasrmn/distributed-cowboys/core/cowboy/usecases"
	"github.com/mindaugasrmn/distributed-cowboys/helpers/kafkautils"
	pbCowboy "github.com/mindaugasrmn/distributed-cowboys/proto/cowboy"
	"google.golang.org/grpc"
)

var (
	name   = flag.String("name", os.Getenv("NAME"), "Cowboy name")
	health = flag.String("health", os.Getenv("HEALTH"), "Cowboy health")
	damage = flag.String("damage", os.Getenv("DAMAGE"), "Cowboy damage")
	port   = flag.String("port", os.Getenv("PORT"), "Cowboy port")
)

// Cowboy runner and server listener starts here
func main() {
	flag.Parse()
	log.Println("Initialized:", name)

	c := cron.New(cron.WithSeconds())
	c.Start()

	mqp, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka-service:9092",
		"acks":              "all",
	})
	if err != nil {
		log.Fatalf("Kafka producer error: %s", err)

	}
	err = kafkautils.CreateTopic("kafka-service:9092", strings.ToLower(*name))
	if err != nil {
		log.Fatalf("Kafka CreateTopic failed, got error: %s", err)

	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to initialize redis client: %s", err)

	}
	list, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	healthInt, err := strconv.Atoi(*health)
	if err != nil {
		log.Fatalf("failed to get health: %v", err)
	}

	damageInt, err := strconv.Atoi(*damage)
	if err != nil {
		log.Fatalf("failed to get damage: %v", err)
	}

	cowboyUse, err := cowboyUsecases.NewCowboyUsecases(rdb, mqp, *name, healthInt, damageInt)
	if err != nil {
		log.Fatalf("failed to init  cowboyUsecases, got error: %s", err.Error())
	}

	grpcSrv := grpc.NewServer()
	server := server{cowboyUse: cowboyUse}
	pbCowboy.RegisterCowboyV1Server(grpcSrv, &server)
	log.Printf("starting grpc server under port %d", port)

	log.Println("started")
	mqc, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               "kafka-service:9092",
		"group.id":                        strings.ToLower(*name) + ".GROUP",
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.partition.eof":            true,
		"auto.offset.reset":               "earliest"})

	if err != nil {
		log.Fatalf("Failed to create consumer: %s\n", err.Error())
		os.Exit(1)
	}
	err = mqc.Subscribe(strings.ToLower(*name), nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %s\n", err.Error())
		os.Exit(1)
	}
	topic := "attenders"
	err = mqp.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(strings.ToLower(*name)),
	}, nil)
	if err != nil {
		log.Fatalf("Failed to sent msg to topic attenders, got error: %s", err)
	}
	go func() {
		err := grpcSrv.Serve(list)
		if err != nil {
			log.Fatalf("failed to start grpc: %s\n", err.Error())
			return
		}
	}()
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
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
				t, err := time.Parse("2006-01-02 15:04:05 -0700 UTC", string(e.Value))

				cronExpression := fmt.Sprintf("%d %d %d %d %s ?", t.Second(), t.Minute(), t.Hour(), t.Day(), strings.ToUpper(t.Month().String()[:3]))
				id, err := c.AddFunc(
					cronExpression,
					func() {
						start(list, grpcSrv, cowboyUse)
					})
				if err != nil {
					log.Fatal(err)
				}
				log.Println("created task id: ", id)

			case kafka.PartitionEOF:
				log.Printf("%% Reached %v\n", e)
			case kafka.Error:
				log.Printf("%% Error: %v\n", e)
			}
		}
	}

	log.Println("Closing consumer and producer")
	mqc.Close()
	mqp.Close()
	grpcSrv.GracefulStop()

}

func start(list net.Listener, grpcSrv *grpc.Server, cowboyUse cowboyUsecases.CowboyUsecases) {
	log.Println("Let the game begins!")
	cowboyUse.Go()
}

type server struct {
	pbCowboy.UnsafeCowboyV1Server
	cowboyUse cowboyUsecases.CowboyUsecases
}

func (s *server) ShootV1(ctx context.Context, req *pbCowboy.ShootRequestV1) (*pbCowboy.ShootResponseV1, error) {
	health := s.cowboyUse.GetHealth()
	if health == 0 {
		return &pbCowboy.ShootResponseV1{Msg: fmt.Sprintf("%s is dead already", s.cowboyUse.GetName())}, nil
	}
	s.cowboyUse.GotDamage(int(req.Damage))
	return &pbCowboy.ShootResponseV1{Msg: fmt.Sprintf("%s got %d damage, he is %s", s.cowboyUse.GetName(), s.cowboyUse.GetDamage(), s.cowboyUse.GetStatus())}, nil
}
