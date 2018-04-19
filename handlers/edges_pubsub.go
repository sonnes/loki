package handlers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/jmoiron/sqlx"
	"gitlab.com/pagalguy/loki/models"
)

const (
	PROJECT_ID string = "prepathon-infrastructure"
	TOPIC_NAME string = "edgestore.edges.sync"
)

type PubsubMessage struct {
	Action    string         `json:"action"`
	Edges     *[]models.Edge `json:"edges"`
	Payload   *[]models.Edge `json:"payload"`
	Timestamp *time.Time     `json:"timestamp"`
}

func StartPubsubListen(db *sqlx.DB) error {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, PROJECT_ID)

	if err != nil {
		return err
	}

	log.Printf("Intialized client: %v", client)

	topic, err := createTopicIfNotExists(client, TOPIC_NAME)

	if err != nil {
		return err
	}

	log.Printf("Intialized topic: %v", topic)

	// create a subscrption
	subscription, err := createSubIfNotExists(client, topic, "edgestore.edges.subscription")

	if err != nil {
		return err
	}

	log.Printf("Listening to subscription: %v", subscription)

	err = subscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {

		var message PubsubMessage
		var err error

		err = json.Unmarshal(m.Data, &message)

		if err != nil {
			log.Printf("Could not parse JSON from message. Ignoring & Acking. Error: %v", err)
			m.Ack()
			return
		}

		if message.Edges == nil {
			log.Print("No edges in payload. Ignoring & Acking")
			m.Ack()
			return
		}

		if message.Action == "/edges/save" {
			err = models.SaveMany(db, message.Edges)
		} else if message.Action == "/edges/delete" {
			err = models.DeleteMany(db, message.Edges)
		}

		if err != nil {
			log.Printf("Error while executing save/delete %v", err)
			m.Nack()
		} else {
			m.Ack()
		}

	})

	if err != nil {
		log.Printf("Error from subscription receive %v", err)
		return err
	}

	return nil
}

func createTopicIfNotExists(client *pubsub.Client, topicName string) (*pubsub.Topic, error) {

	ctx := context.Background()

	// Create a topic to subscribe to.
	t := client.Topic(topicName)

	ok, err := t.Exists(ctx)

	if err != nil {
		return nil, err
	}

	if ok {
		return t, nil
	}

	t, err = client.CreateTopic(ctx, topicName)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func createSubIfNotExists(client *pubsub.Client, topic *pubsub.Topic, subName string) (*pubsub.Subscription, error) {

	ctx := context.Background()

	sub := client.Subscription(subName)

	ok, err := sub.Exists(ctx)

	if err != nil {
		return nil, err
	}

	if ok {
		return sub, nil
	}

	sub, err = client.CreateSubscription(ctx, subName, pubsub.SubscriptionConfig{
		Topic:       topic,
		AckDeadline: 20 * time.Second,
	})

	if err != nil {
		return nil, err
	}

	return sub, nil
}
