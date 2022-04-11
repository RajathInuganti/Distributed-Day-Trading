package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"

	"github.com/streadway/amqp"
)

func HandleConn(conn net.Conn, queue string, ch *amqp.Channel, responses *map[string]net.Conn) {

	var body Command

	for {
		msgsize := make([]byte, 8)
		_, err := conn.Read(msgsize)
		if err != nil && err != io.EOF {
			log.Printf("error while reading: %+v\n", err)
		}
		size := int64(binary.LittleEndian.Uint64(msgsize))

		message := make([]byte, size)
		_, err = conn.Read(message)
		if err != nil && err != io.EOF {
			log.Printf("error while reading: %+v\n", err)
		}

		err = json.Unmarshal(message, &body)
		failOnError("Failed to unmarshal JSON", err)

		CorrelationId := body.Username
		(*responses)[CorrelationId] = conn

		Publish(ch, queue, message, CorrelationId)
	}
}

func Publish(ch *amqp.Channel, queue string, command []byte, CorrelationId string) {
	err := ch.Publish(
		"",
		"server",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: CorrelationId,
			ReplyTo:       queue,
			Body:          command,
		})
	failOnError("Failed to publish a message", err)
}

func startQueueService(ch *amqp.Channel, queue string, responses *map[string]net.Conn) {

	q, err := ch.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	failOnError("Failed to declare a queue", err)

	messages, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError("Failed to register a consumer", err)

	for message := range messages {
		conn := (*responses)[message.CorrelationId]
		conn.Write(message.Body)
	}
}

func main() {

	containerID := os.Getenv("HOSTNAME")
	responses := make(map[string]net.Conn)

	ch := setupChannel()
	go startQueueService(ch, containerID, &responses)

	server, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	log.Println("Listening on localhost:8080")

	for {
		conn, err := server.Accept()
		if err != nil {
			panic(err)
		}

		go HandleConn(conn, containerID, ch, &responses)
	}

}

func setupChannel() *amqp.Channel {

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

func failOnError(message string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", message, err)
	}
}
