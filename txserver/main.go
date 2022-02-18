package main

import (
	"log"
	"context"

	"github.com/streadway/amqp"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var collection *mongo.Collection

func failOnError(message string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", message, err)
	}
}

func consume(ch *amqp.Channel) {

	q, err := ch.QueueDeclare(
		"rpc_queue", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	failOnError("Failed to declare a queue", err)

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError("Failed to set QoS", err)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError("Failed to register a consumer", err)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			command := string(d.Body)
			failOnError("Failed to convert body to integer", err)

			log.Printf("Received command: %s", command)
			// need to called handler from here to handle the various commands
			//parseStruct()

			err = ch.Publish(
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(command),
				})
			failOnError("Failed to publish a message", err)

			err := d.Ack(false)
			failOnError("Failed to Acknowledge message", err)
		}
	}()

	log.Printf(" [*] Awaiting RPC requests")
	<-forever

}

func main() {
	ch := setup()
	setupDB()
	//addtoDb //For testing purposes 
	consume(ch)
}

//testing purposes
func addtoDB() {

	doc := bson.D{{"user", "sadfd23sfds"}, {"account_balance", "98742"}}

	result, err := collection.InsertOne(context.TODO(), doc)

	if err != nil {
		panic(err)
	}

	log.Printf("Inserted document with _id: %v\n", result.InsertedID)
}

func setupDB() {

	const uri = "mongodb://mongodb_container:27017/?maxPoolSize=20&w=majority"
	var client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))	

	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	log.Printf("Successfully connected and pinged.")

	collection = client.Database("test").Collection("transactions")

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
