package messaging

import (
	"context"
	"time"

	"github.com/byvko-dev/am-core/logs"
	amqp "github.com/rabbitmq/amqp091-go"
)

var globalConnection *amqp.Connection
var cacheUpdatesChannel *amqp.Channel

const QueueCacheUpdates = "cache_updates"

func Connect(connString string) (func() error, error) {
	conn, err := amqp.Dial(connString)
	if err != nil {
		return nil, err
	}
	globalConnection = conn

	ch, err := globalConnection.Channel()
	if err != nil {
		return nil, err
	}
	cacheUpdatesChannel = ch

	_, err = ch.QueueDeclare(
		QueueCacheUpdates, // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return nil, err
	}
	return conn.Close, nil
}

func SendQueueMessage(queue string, payload []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return cacheUpdatesChannel.PublishWithContext(ctx,
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         payload,
		})

}

func SubscribeToQueue(queue string, handler func([]byte) error, concurrency int, cancel chan int) error {
	msgs, err := cacheUpdatesChannel.Consume(
		queue, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	limiter := make(chan int, concurrency)
	go func() {
		for d := range msgs {
			limiter <- 1
			go func(d amqp.Delivery) {
				defer func() {
					<-limiter
				}()

				logs.Info("Received a message")

				err := handler(d.Body)
				if err != nil {
					logs.Error("Error while handling message: %s", err)
					d.Nack(false, true)
					time.Sleep(5 * time.Second)
					return
				}
				d.Ack(false)
			}(d)
		}
	}()
	<-cancel
	return nil
}
