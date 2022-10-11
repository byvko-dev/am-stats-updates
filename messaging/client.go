package messaging

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/byvko-dev/am-core/logs"
)

type Client struct {
	client *pubsub.Client
	topic  *pubsub.Topic
	ctx    context.Context
}

func NewClient(project, topic string) (*Client, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("pubsub: NewClient: %v", err)
	}
	t := client.Topic(topic)
	return &Client{
		client: client,
		topic:  t,
		ctx:    ctx,
	}, nil
}

func (s *Client) Close() error {
	s.topic.Stop()
	return s.client.Close()
}

// Publishes a message to the topic
func (s *Client) Publish(message []byte, attributes map[string]string) (string, error) {
	result := s.topic.Publish(s.ctx, &pubsub.Message{
		Attributes: attributes,
		Data:       message,
	})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(s.ctx)
	if err != nil {
		return "", fmt.Errorf("pubsub: result.Get: %v", err)
	}
	fmt.Printf("Published a message; msg ID: %v\n", id)
	return id, nil
}

// Subscribe to the topic with timeout
func (s *Client) Subscribe(subscription string, timeout time.Duration, callback func([]byte) error, onExit func()) error {
	sub := s.client.Subscription(subscription)
	ctx, cancel := context.WithCancel(context.Background())

	// Start a timer to stop the subscriber after timeout
	timer := time.NewTimer(timeout)
	go func() {
		<-timer.C
		cancel()
		onExit()
	}()

	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		logs.Info("Received message: %v", string(msg.ID))

		timer.Reset(timeout) // Reset the timer
		err := callback(msg.Data)
		if err != nil {
			fmt.Printf("message %v will be nacked, callback returned an error: %v", msg.ID, err)
			msg.Nack()
		} else {
			msg.Ack()
		}
	})
	if err != nil {
		return fmt.Errorf("pubsub: Receive: %v", err)
	}
	return nil
}
