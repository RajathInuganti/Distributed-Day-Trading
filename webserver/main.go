package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/streadway/amqp"
)

type commandHandler struct {
	m         *sync.Mutex
	responses *map[string]string
	ch        *amqp.Channel
	queue     string
}

func (c commandHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {

		err := request.ParseForm()
		failOnError("Error while Parsing POST request form", err)

		command := request.Form.Get("command")
		txNum := request.Form.Get("txNum")

		Publish(c.ch, c.queue, command, txNum)
	}
}

func Publish(ch *amqp.Channel, queue string, command string, txNum string) {
	err := ch.Publish(
		"",
		"rpc_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: txNum,
			ReplyTo:       queue,
			Body:          []byte(command),
		})
	failOnError("Failed to publish a message", err)
}

func startQueueService(ch *amqp.Channel, queue string, responses *map[string]string) {

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

	name := "client"
	responses := make(map[string]string)
	var m sync.Mutex

	ch := setupChannel()
	startQueueService(ch, name, &responses)

	handler := commandHandler{m: &m, responses: &responses, ch: ch, queue: name}

	log.Println("Starting server")

	http.Handle("/transaction", handler)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))

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
