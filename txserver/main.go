package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// a global client that can be used across the package
var client *mongo.Client

func failOnError(message string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", message, err)
	}
}

func consume(ctx *context.Context, ch *amqp.Channel) {

	q, err := ch.QueueDeclare(
		"server", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	failOnError("Failed to declare a queue", err)

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError("Failed to set QoS", err)

	messages, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError("Failed to register a consumer", err)

	for message := range messages {
		// need to called handler from here to handle the various commands
		response := handle(ctx, message.Body)

		msgBody, err := json.Marshal(response)
		failOnError("Failed to marshal message body", err)

		err = ch.Publish(
			"",              // exchange
			message.ReplyTo, // routing key
			false,           // mandatory
			false,           // immediate
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: message.CorrelationId,
				Body:          msgBody,
			})
		failOnError("Failed to publish a message", err)

		err = message.Ack(false)
		failOnError("Failed to Acknowledge message", err)
	}
}

func main() {
	ch := setup()
	var cancel context.CancelFunc
	ctx := context.Background()
	client, cancel = setupDB(ctx)
	consume(&ctx, ch)
	cancel()
}

func setup() *amqp.Channel {

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}

	return ch
}

func setupDB(ctx context.Context) (*mongo.Client, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	failOnError("Error while connecting to MongoDB", err)

	//create Indexes here
	Accounts := mongoClient.Database("test").Collection("Accounts")
	model := mongo.IndexModel{
		Keys:    bson.M{"username": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err = Accounts.Indexes().CreateOne(ctx, model)
	failOnError("Account index creation with username failed", err)

	_ = mongoClient.Database("test").Collection("Events")

	_ = mongoClient.Database("test").Collection("Transactions")

	return mongoClient, cancel
}
