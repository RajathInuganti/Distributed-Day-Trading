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
		request.ParseForm()
		command := request.Form.Get("command")
		txNum := request.Form.Get("txNum")

		Publish(c.ch, c.queue, command, txNum)
	}
}

func Publish(ch *amqp.Channel, queue string, command string, txNum string) {
	if err := ch.Publish(
		"",
		"rpc_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: txNum,
			ReplyTo:       queue,
			Body:          []byte(command),
		}); err != nil {
		log.Fatalf("Failed to publish a message: %s", err)
	}
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

	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	messages, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		log.Fatalf("Failed to register a consumer: %s", err)
	}

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
	http.ListenAndServe("localhost:8080", nil)

}

func setupChannel() *amqp.Channel {

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
