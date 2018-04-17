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
	Action    string           `json:"action"`
	Payload   *json.RawMessage `json:"payload"`
	Timestamp *time.Time       `json:"timestamp"`
}

func StartPubsubListen(db *sqlx.DB) error {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, PROJECT_ID)

	if err != nil {
		return err
	}

	topic, err := createTopicIfNotExists(client, TOPIC_NAME)

	// create a subscrption
	subscription, err := createSubIfNotExists(client, topic, "edgestore.edges.subscription")

	if err != nil {
		return err
	}

	err = subscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {

		var message PubsubMessage

		err = json.Unmarshal(m.Data, &message)

		if err != nil {
			log.Print(err)
			log.Fatalf("Could not process data: %v", m.Data)
			m.Ack()
		} else {

			err := HandlePubsubMessage(db, &message)

			if err != nil {
				log.Print(err)
				log.Fatalf("Could not handle data: %v", message)
				m.Ack()
			} else {
				m.Ack()
			}
		}

	})

	if err != nil {
		return err
	}

	return nil
}

func HandlePubsubMessage(db *sqlx.DB, message *PubsubMessage) error {

	var jsonBody EdgesListRequest

	err := json.Unmarshal(*message.Payload, &jsonBody)

	if err != nil {
		return err
	}

	if message.Action == "/edges/save" {
		log.Printf("Saving %d edges", len(*jsonBody.Edges))
		err = models.SaveMany(db, jsonBody.Edges)
	} else if message.Action == "/edges/delete" {
		log.Printf("Deleting %d edges", len(*jsonBody.Edges))
		err = models.DeleteMany(db, jsonBody.Edges)
	}

	if err != nil {
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
