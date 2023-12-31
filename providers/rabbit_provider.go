package providers

import (
	"example/web-service-gin/utils"

	logger "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var (
	rabbitUrl = "amqp://guest:guest@rabbitmq:5672/"

	balanceServiceDeclaredExMsg   = "BalanceService declared ex: %s"
	balanceServiceCreatedQueueMsg = "BalanceService created queue: %s in ex: %s with rk: %s"
	balanceServiceSentMsg         = "BalanceService sent message to ex: %s with rk: %s"
)

type RabbitProvider struct {
	connection *amqp.Connection
}

func NewRabbitProvider() *RabbitProvider {
	conn, err := amqp.Dial(rabbitUrl)

	utils.CheckErrorWithPanic(err)
	rabbitProvider := &RabbitProvider{connection: conn}

	return rabbitProvider
}

func (r *RabbitProvider) getNewChannel() *amqp.Channel {
	ch, err := r.connection.Channel()
	utils.CheckErrorWithPanic(err)

	return ch
}

func (r *RabbitProvider) GetQueueConsumer(exName string, rk string, queueName string) (<-chan amqp.Delivery, *amqp.Channel) {
	ch := r.getNewChannel()
	_, err := ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	utils.CheckErrorWithPanic(err)

	err = ch.QueueBind(
		queueName,
		rk,
		exName,
		false,
		nil,
	)
	utils.CheckErrorWithPanic(err)

	msgs, err := ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	utils.CheckErrorWithPanic(err)

	logger.Infof(balanceServiceCreatedQueueMsg, queueName, exName, rk)
	return msgs, ch
}

func (r *RabbitProvider) DeclareExchange(exName string) {
	err := r.getNewChannel().ExchangeDeclare(
		exName,
		"topic",
		false,
		false,
		false,
		false,
		nil,
	)
	utils.CheckErrorWithPanic(err)
	logger.Infof(balanceServiceDeclaredExMsg, exName)
}

func (r *RabbitProvider) SendMessage(exName string, rk string, message []byte) {
	err := r.getNewChannel().Publish(
		exName,
		rk,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	utils.CheckErrorWithPanic(err)
	logger.Infof(balanceServiceSentMsg, exName, rk)
}

func (r *RabbitProvider) RunListener(msgs <-chan amqp.Delivery, ch *amqp.Channel, BalanceComponentFunc func([]byte)) {
	defer ch.Close()

	forever := make(chan bool)
	go func() {
		for msg := range msgs {
			BalanceComponentFunc(msg.Body)
		}
	}()

	<-forever
}
