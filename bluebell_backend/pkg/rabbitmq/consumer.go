package rabbitmq

import (
	"bluebell_backend/models"
	"bluebell_backend/pkg/email"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func Consumer() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"email_queue", // 与发送发布的队列相同
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	// 监听队列中的消息
	msgs, err := ch.Consume(
		q.Name,
		"",
		true, // 处理完消息后显示发送ACK，避免消息丢失
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	// 异步处理消息
	go func() {
		for d := range msgs {
			var Ed models.RegisterEmailData
			// JSON解码
			err := json.Unmarshal(d.Body, &Ed)
			if err != nil {
				log.Printf("Error decoding JSON: %v\n", err)
				continue
			}
			log.Printf("Sending email for user: %s", Ed.UserName)
			// 处理邮件发送的逻辑
			err = email.SendEmail(&Ed)
			if err != nil {
				log.Printf("Error sending email: %v\n", err)
			}
		}
	}()

	<-forever // 确保消费者持续运行，不会退出
}
