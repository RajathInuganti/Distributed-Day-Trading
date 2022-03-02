package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/streadway/amqp"
)

var txCount int = 0

type commandHandler struct {
	queue     string
	ch        *amqp.Channel
	responses *map[string]chan []byte
}

func (handler *commandHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	var body Command
	if request.Method != "POST" {
		return
	}

	err := json.NewDecoder(request.Body).Decode(&body)
	failOnError("Failed to read HTTP request body", err)

	command, err := json.Marshal(body)
	failOnError("Failed to unmarshal HTTP request body", err)

	txCount = txCount + 1
	CorrelationId := strconv.Itoa(txCount)

	channel := make(chan []byte)
	(*handler.responses)[CorrelationId] = channel

	Publish(handler.ch, handler.queue, command, CorrelationId)

	response := <-channel
	_, err = writer.Write(response)
	if err != nil {
		log.Printf("Unable to write response: %s. Error: %s\n", string(response), err)
	}
	log.Printf("txCount : %d\n", txCount)
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

func startQueueService(ch *amqp.Channel, queue string, responses *map[string]chan []byte) {

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
		channel := (*responses)[message.CorrelationId]
		channel <- message.Body
	}
}

func main() {

	containerID := os.Getenv("HOSTNAME")
	responses := make(map[string]chan []byte)

	ch := setupChannel()
	go startQueueService(ch, containerID, &responses)

	handler := commandHandler{responses: &responses, ch: ch, queue: containerID}

	http.Handle("/transaction", &handler)
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
