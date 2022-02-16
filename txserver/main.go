package main

import (
	"log"

	"github.com/streadway/amqp"
)

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

			d.Ack(false)
		}
	}()

	log.Printf(" [*] Awaiting RPC requests")
	<-forever

}

func main() {
	ch := setup()
	consume(ch)
}

func setup() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
		conn.Close()
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
		ch.Close()
	}

	return ch
}
