package kafkautils

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func CreateTopic(bootstrapServers, topic string) error {

	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
	})

	if err != nil {
		return fmt.Errorf("Failed to create Admin client: %s", err)
	}

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create topics on cluster.
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDuration, err := time.ParseDuration("60s")
	if err != nil {
		return fmt.Errorf("time.ParseDuration(60s)")
	}

	results, err := adminClient.CreateTopics(ctx,
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1}},
		kafka.SetAdminOperationTimeout(maxDuration))

	if err != nil {

		return fmt.Errorf("Problem during the topic creation: %v", err)
	}

	// Check for specific topic errors
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError &&
			result.Error.Code() != kafka.ErrTopicAlreadyExists {

			return fmt.Errorf("Topic creation failed for %s: %v",
				result.Topic, result.Error.String())
		}
	}

	adminClient.Close()
	return nil

}
