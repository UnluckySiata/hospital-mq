package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	HOSPITAL_EXCHANGE = "hospital"
)

func main() {
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

	err = ch.QueueBind(q.Name, "tech.*", HOSPITAL_EXCHANGE, false, nil)
	if err != nil {
		fmt.Printf("Failed to bind queue: %v\n", err)
		return
	}

	err = ch.QueueBind(q.Name, "doc.*", HOSPITAL_EXCHANGE, false, nil)
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
	go handleInfo(msgs)

	reader := bufio.NewReader(os.Stdin)
	for {
		read, err := reader.ReadBytes('\n')

		if err == io.EOF {
			fmt.Println("exiting...")
			break
		}

		info := string(read[:len(read)-1])

		err = ch.Publish(
			HOSPITAL_EXCHANGE, // exchange
			"info",            // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(info),
			})

		if err != nil {
			fmt.Printf("Failed to publish: %v\n", err)
		}
	}
}

func handleInfo(msgs <-chan amqp.Delivery) {
	for d := range msgs {
		log.Printf("%s\n", d.Body)
		d.Ack(false)
	}
}
