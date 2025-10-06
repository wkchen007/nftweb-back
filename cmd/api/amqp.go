package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (app *application) connectToAmqp() (*amqp.Connection, error) {
	conn, err := amqp.Dial(app.amqpURL)
	if err != nil {
		log.Printf("connect RabbitMQ failed: %v", err)
		return nil, err
	}

	log.Println("Connected to RabbitMQ!")
	return conn, nil
}
