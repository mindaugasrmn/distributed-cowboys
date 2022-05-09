package cowboy

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis/v8"
	pbCowboy "github.com/mindaugasrmn/distributed-cowboys/proto/cowboy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	DefaultNamespace = "default"
)

type cowboyUsecases struct {
	rdb        *redis.Client
	mqp        *kafka.Producer
	randVictim *rand.Rand
	Name       string
	Health     int
	InitHealth int
	Damage     int
	Status     string
	logdAddr   string
}

func NewCowboyUsecases(rdb *redis.Client, mqp *kafka.Producer, name string, health, damage int) (CowboyUsecases, error) {
	return &cowboyUsecases{
		rdb:        rdb,
		mqp:        mqp,
		Health:     health,
		randVictim: rand.New(rand.NewSource(time.Now().UnixNano())),
		Damage:     damage,
		Name:       name,
		Status:     "alive",
		InitHealth: health,
	}, nil
}

type CowboyUsecases interface {
	Go()
	Shoot(target string) (string, string, error)
	GotDamage(damage int)
	GetHealth() int
	GetName() string
	GetDamage() int
	GetStatus() string
}

func (s *cowboyUsecases) Go() {
	for {
		if s.Health <= 0 {
			log.Println("Im dead")
			break
		}

		itemIdx, target, err := getVictim(s)
		if itemIdx != nil && target != nil && err == nil {
			msg, _, err := s.Shoot(*target)
			if err != nil {
				log.Printf("failed to shoot: %v \n", err)
				continue
			}
			err = s.log(fmt.Sprintf("%s /= - %s", s.GetName(), msg))
			if err != nil {
				log.Printf("%v \n", err)
			}
		}
		if itemIdx == nil && target == nil && err == nil {
			err := s.log(fmt.Sprintf("all dead except %s. %s is the winner!", s.Name, s.Name))
			if err != nil {
				log.Printf("%v \n", err)
				continue
			}
			break
		}
		log.Printf("resting\n")
		time.Sleep(time.Second)
	}
	return
}

func (s *cowboyUsecases) Shoot(target string) (string, string, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", "", err
	}
	defer conn.Close()
	c := pbCowboy.NewCowboyV1Client(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := c.ShootV1(ctx, &pbCowboy.ShootRequestV1{Damage: uint64(s.Damage)})
	if err != nil {
		return "", "", err
	}
	return resp.Msg, resp.Status, nil
}

func (s *cowboyUsecases) GotDamage(damage int) {
	actualHealth := s.Health - damage
	if s.InitHealth/2 < actualHealth && actualHealth > 0 {
		s.Status = "still alive"
	} else if s.InitHealth/2 > actualHealth && actualHealth > 0 {
		s.Status = "bearly alive"
	} else if actualHealth <= 0 {
		log.Println("Im dying, i have health left", actualHealth)
		actualHealth = 0
		_, err := s.rdb.Del(context.Background(), strings.ToLower(s.GetName())).Result()
		if err != nil {
			log.Println("failed to remove attendee, got error:", err)
		}
		s.Health = actualHealth
		s.Status = "dead"
		return
	}
	s.Health = actualHealth
	log.Println("I", s.Name, " have", s.Health, " health left.")
	return
}

func (s *cowboyUsecases) GetStatus() string {
	return s.Status
}

func (s *cowboyUsecases) GetHealth() int {
	return s.Health
}

func (s *cowboyUsecases) GetName() string {
	return s.Name
}

func (s *cowboyUsecases) GetDamage() int {
	return s.Damage
}

func (s *cowboyUsecases) log(logEntry string) error {
	log.Println("publishing to log topic:", logEntry)
	topic := "log"
	err := s.mqp.Produce(&kafka.Message{

		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(logEntry),
	}, nil)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("0306")
	}
	return nil
}

func getVictim(s *cowboyUsecases) (*int, *string, error) {
	//not very good idea for big db
	keys, err := s.rdb.Keys(context.Background(), "*").Result()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get keys from redis db, got error: %s", err.Error())
	}
	if len(keys) == 1 && keys[0] == strings.ToLower(s.Name) {
		return nil, nil, nil
	}
	keys = remove(keys, strings.ToLower(s.Name))
	index := s.randVictim.Intn(len(keys))
	target, err := s.rdb.Get(context.Background(), keys[index]).Result()
	log.Printf("got target %s \n", target)
	return &index, &target, nil
}

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			log.Println("retrying after error:", err)
			time.Sleep(sleep)
			sleep *= 2
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
