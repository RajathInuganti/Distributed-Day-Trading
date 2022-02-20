package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

func failOnError(message string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", message, err)
	}
}

func consume(ch *amqp.Channel) {

	command := new(Command)

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
		err := json.Unmarshal(message.Body, &command)
		failOnError("Failed to unmarshal message body message", err)

		// need to called handler from here to handle the various commands
		handle(command)

		msgBody, err := json.Marshal(command)
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
	//setupDB()
	//addtoDb //For testing purposes
	consume(ch)
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
