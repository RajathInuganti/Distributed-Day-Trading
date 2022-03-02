package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/streadway/amqp"
)

type commandHandler struct {
	txCount   int
	queue     string
	lock      *sync.Mutex
	ch        *amqp.Channel
	responses *map[string][]byte
}

func (handler commandHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	var body Command
	if request.Method != "POST" {
		return
	}

	err := json.NewDecoder(request.Body).Decode(&body)
	failOnError("Failed to read HTTP request body", err)

	command, err := json.Marshal(body)
	failOnError("Failed to unmarshal HTTP request body", err)

	handler.txCount = handler.txCount + 1
	Publish(handler.ch, handler.queue, command, strconv.Itoa(handler.txCount))

	value, ok := (*handler.responses)[strconv.Itoa(handler.txCount)]
	for !ok {
		value, ok = (*handler.responses)[strconv.Itoa(handler.txCount)]
	}
	_, err = writer.Write(value)
	if err != nil {
		log.Printf("Error writing response: %s", err)
	}

}

func Publish(ch *amqp.Channel, queue string, command []byte, txNum string) {
	err := ch.Publish(
		"",
		"server",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: txNum,
			ReplyTo:       queue,
			Body:          command,
		})
	failOnError("Failed to publish a message", err)
}

func startQueueService(ch *amqp.Channel, queue string, responses *map[string][]byte, lock *sync.Mutex) {

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
		(*responses)[message.CorrelationId] = message.Body
	}
}

func main() {

	containerID := os.Getenv("HOSTNAME")
	responses := make(map[string][]byte)
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
