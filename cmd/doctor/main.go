package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	HOSPITAL_EXCHANGE = "hospital"
)

var (
	doc = flag.String("n", "doc", "doc")
)

func main() {
	flag.Parse()
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	if err != nil {
		fmt.Printf("Failed to connect to RabbitMQ: %v\n", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Failed to create channel: %v\n", err)
		return
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(HOSPITAL_EXCHANGE, "topic", true, false, false, false, nil)
	if err != nil {
		fmt.Printf("Failed to declare exchange: %v\n", err)
		return
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		fmt.Printf("Failed to declare queue: %v\n", err)
		return
	}
	key := fmt.Sprintf("doc.%s", *doc)
	err = ch.QueueBind(q.Name, key, HOSPITAL_EXCHANGE, false, nil)
	if err != nil {
		fmt.Printf("Failed to bind queue: %v\n", err)
		return
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		fmt.Printf("Failed to consume: %v\n", err)
		return
	}
	go handleFinished(msgs)

	q, err = ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		fmt.Printf("Failed to declare queue: %v\n", err)
		return
	}
	err = ch.QueueBind(q.Name, "info", HOSPITAL_EXCHANGE, false, nil)
	if err != nil {
		fmt.Printf("Failed to bind queue: %v\n", err)
		return
	}

	msgs, err = ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		fmt.Printf("Failed to consume: %v\n", err)
		return
	}
	go handleInfo(msgs)

	fmt.Printf("Welcome doctor %s!\n", *doc)
	fmt.Println("Order examination:\n[variant] [patient name]")
	reader := bufio.NewReader(os.Stdin)
	for {
		read, err := reader.ReadBytes('\n')

		if err == io.EOF {
			fmt.Println("exiting...")
			break
		}

		info := strings.SplitN(string(read[:len(read)-1]), " ", 2)
		if len(info) != 2 {
			fmt.Println("Bad request")
			continue
		}
		variant := info[0]
		patient := info[1]

		key := fmt.Sprintf("tech.%s", variant)
		body := fmt.Sprintf("%s %s %s", *doc, variant, patient)

		err = ch.Publish(
			HOSPITAL_EXCHANGE, // exchange
			key,               // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})

		if err != nil {
			fmt.Printf("Failed to publish: %v\n", err)
		}
	}
}

func handleFinished(msgs <-chan amqp.Delivery) {
	for d := range msgs {
		fmt.Printf("Incoming: %s\n", d.Body)
		d.Ack(false)
	}
}

func handleInfo(msgs <-chan amqp.Delivery) {
	for d := range msgs {
		fmt.Printf("INFO: %s\n", d.Body)
		d.Ack(false)
	}
}
