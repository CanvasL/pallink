package mq

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type FanoutConfig struct {
	URL      string `json:"url"`
	Exchange string `json:"exchange"`
}

type FanoutPublisher struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	exchange string
}

type FanoutSubscriber struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	queueName string
}

func NewFanoutPublisher(cfg FanoutConfig) (*FanoutPublisher, error) {
	if cfg.URL == "" || cfg.Exchange == "" {
		return nil, errors.New("rabbitmq url/exchange required")
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
	if err := ch.ExchangeDeclare(cfg.Exchange, "fanout", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return &FanoutPublisher{
		conn:     conn,
		ch:       ch,
		exchange: cfg.Exchange,
	}, nil
}

func (p *FanoutPublisher) PublishJSON(ctx context.Context, body any) error {
	if p == nil || p.ch == nil {
		return errors.New("fanout publisher not initialized")
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.ch.PublishWithContext(ctx, p.exchange, "", false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         data,
		DeliveryMode: amqp.Persistent,
	})
}

func (p *FanoutPublisher) Close() error {
	if p == nil {
		return nil
	}
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

func NewFanoutSubscriber(cfg FanoutConfig) (*FanoutSubscriber, error) {
	if cfg.URL == "" || cfg.Exchange == "" {
		return nil, errors.New("rabbitmq url/exchange required")
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
	if err := ch.ExchangeDeclare(cfg.Exchange, "fanout", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	if err := ch.QueueBind(q.Name, "", cfg.Exchange, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return &FanoutSubscriber{
		conn:      conn,
		ch:        ch,
		queueName: q.Name,
	}, nil
}

func (s *FanoutSubscriber) Consume() (<-chan amqp.Delivery, error) {
	if s == nil || s.ch == nil {
		return nil, errors.New("fanout subscriber not initialized")
	}
	if err := s.ch.Qos(50, 0, false); err != nil {
		return nil, err
	}
	return s.ch.Consume(s.queueName, "", false, true, false, false, nil)
}

func (s *FanoutSubscriber) Close() error {
	if s == nil {
		return nil
	}
	if s.ch != nil {
		_ = s.ch.Close()
	}
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func dialWithRetry(url string) (*amqp.Connection, error) {
	var (
		conn *amqp.Connection
		err  error
	)
	for i := 0; i < 30; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			return conn, nil
		}
		time.Sleep(time.Second)
	}
	return nil, err
}
