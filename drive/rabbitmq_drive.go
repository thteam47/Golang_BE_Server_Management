package drive

import (
	"fmt"
	"log"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	RBit *amqp.Connection
}

var Rabbit = &RabbitMQ{}

func ConnectRabbitMQ(userrb string, passrb string, hostrb string, portrb string) *RabbitMQ {
	connAddr := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/", userrb, passrb, hostrb, portrb,
	)
	rabbitmq, err := amqp.Dial(connAddr)
	if err != nil {
		log.Fatalf("Error creating the rabbitmq: %s", err)
	}
	Rabbit.RBit = rabbitmq
	return Rabbit
}
