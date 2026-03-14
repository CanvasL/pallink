package mq

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	URL   string `json:"url"`
	Queue string `json:"queue"`
}

type Client struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.URL == "" || cfg.Queue == "" {
		return nil, errors.New("rabbitmq url/queue required")
	}
	conn, err := dialWithRetry(cfg.URL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	q, err := ch.QueueDeclare(
		cfg.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	return &Client{conn: conn, ch: ch, queue: q}, nil
}

func (c *Client) PublishJSON(ctx context.Context, body any) error {
	if c == nil || c.ch == nil {
		return errors.New("rabbitmq not initialized")
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.ch.PublishWithContext(
		ctx,
		"",
		c.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         data,
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (c *Client) Consume() (<-chan amqp.Delivery, error) {
	if c == nil || c.ch == nil {
		return nil, errors.New("rabbitmq not initialized")
	}
	if err := c.ch.Qos(1, 0, false); err != nil {
		return nil, err
	}
	return c.ch.Consume(
		c.queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
