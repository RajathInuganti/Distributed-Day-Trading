package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/streadway/amqp"
)

func HandleConn(conn net.Conn, queue string, ch *amqp.Channel, responses *map[string]net.Conn,
	lock *sync.Mutex) {

	var body Command

	for {
		msgsize := make([]byte, 8)
		_, err := conn.Read(msgsize)
		if err != nil && err != io.EOF {
			log.Printf("error while reading: %+v\n", err)
		}
		size := int64(binary.LittleEndian.Uint64(msgsize))
		if size == 0 {
			err = conn.Close()
			failOnError("Could not close connection", err)
			return
		}

		message := make([]byte, size)
		_, err = conn.Read(message)
		if err != nil && err != io.EOF {
			log.Printf("error while reading: %+v\n", err)
		}

		err = json.Unmarshal(message, &body)
		failOnError("Failed to unmarshal JSON", err)

		CorrelationId := body.Username
		lock.Lock()
		(*responses)[CorrelationId] = conn
		lock.Unlock()

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

func startQueueService(ch *amqp.Channel, queue string, responses *map[string]net.Conn,
	lock *sync.Mutex) {

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
		lock.Lock()
		conn := (*responses)[message.CorrelationId]
		lock.Unlock()
		_, err = conn.Write(message.Body)
		failOnError("Failed to send response", err)
	}
}

func main() {

	containerID := os.Getenv("HOSTNAME")
	responses := make(map[string]net.Conn)

	var lock sync.Mutex
	ch := setupChannel()
	go startQueueService(ch, containerID, &responses, &lock)

	server, err := net.Listen("tcp", os.Getenv("WEBSERVER_URL"))
	if err != nil {
		panic(err)
	}

	log.Println("Listening on localhost:8080")

	for {
		conn, err := server.Accept()
		if err != nil {
			panic(err)
		}
		go HandleConn(conn, containerID, ch, &responses, &lock)
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
