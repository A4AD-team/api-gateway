package broker

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *amqp.Config
}

func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Configure QoS
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, err
	}

	return &RabbitMQClient{
		conn:    conn,
		channel: ch,
	}, nil
}

func (c *RabbitMQClient) DeclareQueue(name string) error {
	_, err := c.channel.QueueDeclare(
		name,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return err
}

func (c *RabbitMQClient) PublishMessage(queue string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	err = c.channel.Publish(
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         data,
			DeliveryMode: amqp.Persistent, // make messages persistent
		})

	return err
}

func (c *RabbitMQClient) ConsumeMessages(queue string) (<-chan amqp.Delivery, error) {
	msgs, err := c.channel.Consume(
		queue,
		"",    // consumer
		false, // auto-ack (false for manual acknowledgment)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	return msgs, err
}

func (c *RabbitMQClient) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
