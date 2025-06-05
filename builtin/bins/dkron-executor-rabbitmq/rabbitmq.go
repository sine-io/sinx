package main

import (
	"encoding/base64"
	"errors"
	"strconv"

	zlog "github.com/rs/zerolog/log"
	"github.com/streadway/amqp"

	sxplugin "github.com/sine-io/sinx/plugin"
	sxproto "github.com/sine-io/sinx/types"
)

// RabbitMQ process publish rabbitmq message when Execute method is called.
type RabbitMQ struct{}

// Execute method of the plugin
// "executor": "rabbitmq",
//
//	"executor_config": {
//			"url": "amqp://guest:guest@localhost:5672/",
//			"queue.name": "test",
//			"queue.create": "true",
//			"queue.durable": "true",
//			"queue.auto_delete": "false",
//			"queue.exclusive": "false",
//			"message.content_type": "application/json",
//			"message.delivery_mode": "2",
//			"message.messageId": "4373732772",
//			"message.body": "{\"key\":\"value\"}"
//			"message.base64Body": "base64encodedBody"
//	}
func (r *RabbitMQ) Execute(args *sxproto.ExecuteRequest, cb sxplugin.StatusHelper) (*sxproto.ExecuteResponse, error) {
	out, err := r.ExecuteImpl(args, cb)
	resp := &sxproto.ExecuteResponse{Output: out}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp, nil
}

// ExecuteImpl do rabbitmq publish
func (r *RabbitMQ) ExecuteImpl(args *sxproto.ExecuteRequest, cb sxplugin.StatusHelper) ([]byte, error) {
	// validate config
	cfg := args.Config
	if cfg == nil {
		return nil, errors.New("RabbitMQ config is empty")
	}

	url := cfg["url"]
	if url == "" {
		return nil, errors.New("RabbitMQ url is empty")
	}

	queueName := cfg["queue.name"]
	if queueName == "" {
		return nil, errors.New("RabbitMQ queue name is empty")
	}

	if cfg["message.body"] != "" && cfg["message.base64Body"] != "" {
		return nil, errors.New("RabbitMQ message.body and message.base64Body are both set")
	}

	// establish connection
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	defer func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			zlog.Error().Err(err).Msg("Failed to close amqp connection")
		}
	}(conn)

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			zlog.Error().Err(err).Msg("Failed to close channel")
		}
	}(ch)

	// create queue if necessary
	if err := createQueueIfNecessary(cfg, queueName, ch); err != nil {
		return nil, err
	}

	// publish message
	if err = publish(cfg, ch); err != nil {
		return nil, err
	}
	return nil, nil
}

func createQueueIfNecessary(cfg map[string]string, queue string, ch *amqp.Channel) error {
	if val, ok := cfg["queue.create"]; !ok || (ok && val == "false") {
		return nil
	}

	durable, _ := strconv.ParseBool(cfg["queue.durable"])
	autoDelete, _ := strconv.ParseBool(cfg["queue.auto_delete"])
	exclusive, _ := strconv.ParseBool(cfg["queue.exclusive"])

	_, err := ch.QueueDeclare(
		queue,
		durable,
		autoDelete,
		exclusive,
		false,
		nil,
	)

	return err
}

func publish(cfg map[string]string, ch *amqp.Channel) error {
	var body []byte
	b64, ok := cfg["message.base64Body"]
	if ok {
		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return err
		}
		body = decoded
	} else {
		stringBody := cfg["message.body"]
		body = []byte(stringBody)
	}

	contentType := cfg["message.content_type"]
	if contentType == "" {
		contentType = "text/plain"
	}
	messageId := cfg["message.messageId"]
	rawDeliveryMode := cfg["message.delivery_mode"]
	if rawDeliveryMode == "" {
		rawDeliveryMode = "0"
	}
	deliveryMode, err := strconv.ParseUint(rawDeliveryMode, 10, 8)
	if err != nil {
		return err
	}
	return ch.Publish(
		"",                // exchange
		cfg["queue.name"], // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType:  contentType,
			Body:         body,
			MessageId:    messageId,
			DeliveryMode: uint8(deliveryMode),
		})
}
