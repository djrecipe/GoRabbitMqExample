package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func sendRPC(command string, key int, value int) (res string, err error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	corrId := randomString(32)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	key_str := strconv.Itoa(key);
	value_str := strconv.Itoa(value);
	reg := []string {command,key_str,value_str}
	body := strings.Join(reg[:],",")

	err = ch.PublishWithContext(ctx,
		"",          // exchange
		"rpc_queue", // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          []byte(body),
		})
	failOnError(err, "Failed to publish a message")

	for d := range msgs {
		if corrId == d.CorrelationId {
			res = string(d.Body)
			break
		}
	}

	return
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	command,value1,value2 := bodyFrom(os.Args)

	log.Printf(" [x] Requesting %s(%d,%d)", command,value1,value2)
	res, err := sendRPC(command,value1,value2)
	failOnError(err, "Failed to handle RPC request")

	log.Printf(" [.] Response: %s", res)
}

func bodyFrom(args []string) (string,int,int) {
	if (len(args) < 3) || os.Args[1] == "" {
		log.Panicf("Invalid args")
	} else if len(args) < 4 {
		command := args[1]
		if(command == "set") {
			log.Panicf("Invalid args")
		}
		key, err := strconv.Atoi(args[2])
		failOnError(err, "Failed to convert key to integer")
		value := -1
		return command,key,value
	} else {
		command := args[1]
		key, err := strconv.Atoi(args[2])
		failOnError(err, "Failed to convert key to integer")
		value, err := strconv.Atoi(args[3])
		failOnError(err, "Failed to convert value to integer")
		return command,key,value
	}
	return "",-1,-1
}
