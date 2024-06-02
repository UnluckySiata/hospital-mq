package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	HOSPITAL_EXCHANGE = "hospital"
)

var (
	variants = []string{}
	hip      = flag.Bool("h", false, "hip")
	knee     = flag.Bool("k", false, "knee")
	elbow    = flag.Bool("e", false, "elbow")
)

func main() {
	flag.Parse()

	if *elbow {
		variants = append(variants, "elbow")
	}
	if *knee {
		variants = append(variants, "knee")
	}
	if *hip {
		variants = append(variants, "hip")
	}

	fmt.Printf("Operation variants: %s\n", strings.Join(variants, ", "))

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

	err = ch.Qos(1, 0, true)
	if err != nil {
		fmt.Printf("Failed to set qos: %v\n", err)
		return
	}

	err = ch.ExchangeDeclare(HOSPITAL_EXCHANGE, "topic", true, false, false, false, nil)
	if err != nil {
		fmt.Printf("Failed to declare exchange: %v\n", err)
		return
	}

	for _, v := range variants {
		q, err := ch.QueueDeclare(
			v,     // name
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
		key := fmt.Sprintf("tech.%s", v)
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
		go handleJobs(ch, v, msgs)
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
	err = ch.QueueBind(q.Name, "info", HOSPITAL_EXCHANGE, false, nil)
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

	var forever chan struct{}
	<-forever
}

func handleInfo(msgs <-chan amqp.Delivery) {
	for d := range msgs {
		fmt.Printf("INFO: %s\n", d.Body)
		d.Ack(false)
	}
}

func handleJobs(ch *amqp.Channel, variant string, msgs <-chan amqp.Delivery) {
	for d := range msgs {
		info := strings.SplitN(string(d.Body), " ", 3)

		if len(info) != 3 {
			d.Ack(false)
			continue
		}

		doc := info[0]
		patient := info[2]

		reply := fmt.Sprintf("%s %s done", patient, variant)
		key := fmt.Sprintf("doc.%s", doc)

		fmt.Printf("New patient: %s\n", patient)
		t := rand.Intn(4) + 1
		time.Sleep(time.Second * time.Duration(t))
		fmt.Printf("Processed: %s after %d secs\n", patient, t)

		err := ch.Publish(
			HOSPITAL_EXCHANGE, // exchange
			key,               // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(reply),
			})

		if err != nil {
			fmt.Printf("Failed to publish: %v\n", err)
		}
		d.Ack(false)
	}
}
