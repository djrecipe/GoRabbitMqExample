package main

import (
	"context"
	"log"
	"strconv"
	"time"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

var pairs map[int]int

func get(key int) int {
	return pairs[key];
}

func set(key int, value int) {
	pairs[key] = value
}

func main() {
	pairs = make(map[int]int)
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"rpc_queue", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var response string
		for d := range msgs {
			var split []string = strings.Split(string(d.Body),",")

			if(strings.Contains(string(d.Body), "get")) {
				key, err := strconv.Atoi(string(split[1]))
				failOnError(err, "Failed to convert key to integer")
				log.Printf("get(%d)", key)
				response = strconv.Itoa(get(key))
			}

			if(strings.Contains(string(d.Body), "set")) {
				key, err := strconv.Atoi(string(split[1]))
				failOnError(err, "Failed to convert key to integer")
				value, err := strconv.Atoi(string(split[2]))
				failOnError(err, "Failed to convert value to integer")
				log.Printf("set(%d, %d)", key, value)
				set(key, value)
				response = "success"
			}

			err = ch.PublishWithContext(ctx,
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(response),
				})
			failOnError(err, "Failed to publish a message")

			d.Ack(false)
		}
	}()

	log.Printf(" [*] Awaiting RPC requests")
	<-forever
}
