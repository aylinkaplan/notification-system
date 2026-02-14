package queue

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	queuePrefix  = "notifications."
	maxPriority  = 10
	priorityHigh = 10
	priorityNorm = 5
	priorityLow  = 1
)

// RabbitMQQueue implements Queue using RabbitMQ with priority queues.
type RabbitMQQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.Mutex
}

// NewRabbitMQQueue connects to RabbitMQ and declares queues (one per channel with priority).
func NewRabbitMQQueue(url string) (*RabbitMQQueue, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	q := &RabbitMQQueue{conn: conn, channel: ch}
	for _, chName := range []string{"sms", "email", "push"} {
		queueName := queuePrefix + chName
		_, err = ch.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			amqp.Table{"x-max-priority": maxPriority},
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, err
		}
	}
	return q, nil
}

func (q *RabbitMQQueue) queueName(channel string) string {
	return queuePrefix + channel
}

func priorityToInt(priority string) uint8 {
	switch priority {
	case "high":
		return priorityHigh
	case "low":
		return priorityLow
	default:
		return priorityNorm
	}
}

func (q *RabbitMQQueue) Push(ctx context.Context, channel string, notificationID uuid.UUID, priority string) error {
	item := QueuedItem{
		NotificationID: notificationID,
		Priority:       priority,
		QueuedAt:       time.Now(),
	}
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.channel.PublishWithContext(ctx,
		"",
		q.queueName(channel),
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Priority:     priorityToInt(priority),
			ContentType:  "application/json",
			Body:         data,
		},
	)
}

func (q *RabbitMQQueue) Pop(ctx context.Context, channel string, timeout time.Duration) (*QueuedItem, error) {
	queueName := q.queueName(channel)
	q.mu.Lock()
	// Use Get for one-shot receive (no long-lived consumer).
	msg, ok, err := q.channel.Get(queueName, false)
	q.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	var item QueuedItem
	if err := json.Unmarshal(msg.Body, &item); err != nil {
		_ = msg.Nack(false, true)
		return nil, err
	}
	_ = msg.Ack(false)
	return &item, nil
}

func (q *RabbitMQQueue) Depth(ctx context.Context, channel string) (int64, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	queueName := q.queueName(channel)
	info, err := q.channel.QueueInspect(queueName)
	if err != nil {
		return 0, err
	}
	return int64(info.Messages), nil
}

func (q *RabbitMQQueue) AllDepths(ctx context.Context) (map[string]int64, error) {
	channels := []string{"sms", "email", "push"}
	depths := make(map[string]int64)
	for _, ch := range channels {
		d, err := q.Depth(ctx, ch)
		if err != nil {
			return nil, err
		}
		depths[ch] = d
	}
	return depths, nil
}

// Close closes the channel and connection.
func (q *RabbitMQQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.channel != nil {
		_ = q.channel.Close()
		q.channel = nil
	}
	if q.conn != nil {
		_ = q.conn.Close()
		q.conn = nil
	}
	return nil
}
