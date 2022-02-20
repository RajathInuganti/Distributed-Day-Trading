package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/streadway/amqp"
)

type commandHandler struct {
	lock      *sync.Mutex
	responses *map[string]string
	ch        *amqp.Channel
	queue     string
	txCount   int
}

func (handler commandHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	var command []byte
	if request.Method == "POST" {

		_, err := request.Body.Read(command)
		failOnError("Failed to read HTTP request body", err)

		handler.txCount = handler.txCount + 1
		Publish(handler.ch, handler.queue, command, strconv.Itoa(handler.txCount))
	}
}

func Publish(ch *amqp.Channel, queue string, command []byte, txNum string) {
	err := ch.Publish(
		"",
		"rpc_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: txNum,
			ReplyTo:       queue,
			Body:          command,
		})
	failOnError("Failed to publish a message", err)

	log.Println("Successfully Published message")
}

func startQueueService(ch *amqp.Channel, queue string, responses *map[string]string, lock *sync.Mutex) {

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
		messageJSON := string(message.Body)
		log.Printf("Received message: %s", messageJSON)
	}
}

func main() {

	containerID := os.Getenv("HOSTNAME")
	responses := make(map[string]string)
	var lock sync.Mutex

	ch := setupChannel()
	go startQueueService(ch, containerID, &responses, &lock)

	handler := commandHandler{lock: &lock, responses: &responses, ch: ch, queue: containerID, txCount: 0}

	http.Handle("/transaction", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))

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
