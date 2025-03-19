package rabbitmq

import (
	"bluebell_backend/models"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

// 辅助函数检查每个amqp调用的返回值
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	// 创建连接 amqp://username:password@host:port/
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// 创建通道，用于后续声明队列和发布消息
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// 声明消息队列
	q, err := ch.QueueDeclare(
		"bell_queue",
		true,  // 持久化队列，即使RabbitMQ重启，队列也不会丢失
		false, // consumer断开后是否delete队列
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	body := "SignUp by email"
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s", body)
}

// PublishEmailTask 邮件任务发布
func PublishEmailTask(Ed *models.RegisterEmailData) error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"email_queue",
		true, // 即使RabbitMQ重启，队列也不会丢失
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	// 将结构体编码为json
	body, err := json.Marshal(Ed)
	if err != nil {
		return err
	}

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent email task for %s", Ed.Email)
	return nil
}
