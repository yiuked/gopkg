package rabbit

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
)

type Option struct {
	Url string `mapstructure:"url" json:"url" yaml:"url"` // "amqp://guest:guest@localhost:5672/"
}

type Client struct {
	opts *Option
	conn *amqp.Connection
}

func NewRabbit(option *Option) (*Client, error) {
	config := amqp.Config{
		Heartbeat: 30 * time.Second,
	}
	conn, err := amqp.DialConfig(option.Url, config)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", "Failed to connect to RabbitMQ", err)
	}
	return &Client{conn: conn, opts: option}, nil
}

// createChannelAndQueueDeclare 创建管道与队列，注意（声明队列时，生产端与消费端的args参数要一样，否则会出现声明冲突问题）
func (c *Client) createChannelAndQueueDeclare(queue string, toDead bool, expire int64) (*amqp.Channel, *amqp.Queue, error) {
	// 每个应用程序都应该为每个并发处理任务创建一个channel。这样做可以让不同的任务之间独立地进行消息传递，避免了竞争和干扰。
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %s", "Failed to open a channel", err)
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %s", "Failed to set channel Qos", err)
	}

	// 创建死信队列（只发布端有该需求）
	var args amqp.Table
	if toDead {
		deadQueue := fmt.Sprintf("dead:%s", queue)
		args = amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": deadQueue,
		}
		// 声明一个死信队列，用于存放由于过期或者其他原因被丢弃的消息
		_, err := ch.QueueDeclare(
			deadQueue,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to declare a dead-letter queue: %v", err)
		}
	}
	if expire > 0 {
		args["x-message-ttl"] = expire * 1000
	}

	// 声明一个队列，如果队列已存在，那么执行ch.QueueDeclare函数仅仅是检查该队列是否存在，并返回相应的队列信息。
	// 不会对现有的队列做任何更改。如果队列不存在，那么执行ch.QueueDeclare函数时，就会创建一个新的队列。
	// 需要注意的是，RabbitMQ中的队列名称是全局唯一的。也就是说，不同的客户端不能重复声明同名的队列。
	// 如果客户端尝试声明一个已经存在的同名队列但参数不同（例如durable和exclusive参数），则会抛出异常。
	// 因此，客户端在使用ch.QueueDeclare函数时需要确保队列名称是唯一的，并且参数要和现有队列一致。否则可能会导致意外错误。
	q, err := ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,  // arguments
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %s", "Failed to declare a queue", err)
	}
	return ch, &q, nil
}

// Publish 发布信息
// expire 超过这个时间没被消费则丢入死信队列（秒）
func (c *Client) Publish(queue string, body []byte, expire int64) error {
	ch, q, err := c.createChannelAndQueueDeclare(queue, true, expire)
	if err != nil {
		return err
	}
	defer ch.Close()

	msg := amqp.Publishing{
		ContentType: "text/plain",
		Body:        body,
	}

	// 设置消息超时时间，超出TTL时间，则会被丢弃
	if expire > 0 {
		headers := amqp.Table{"expiration": expire * 1000}
		msg.Headers = headers
	}

	// 将一个payload（二进制数据），按照指定的规则（使用特定的 routing key 等），发布到 RabbitMQ 中。
	// 这样，其他订阅了对应 exchange 上相同 routing key 的队列的消费者就可以接收到该消息并处理。
	// 当 `exchange` 参数为空字符串时，即表示将消息发送到默认的 exchange 上。
	// 在 RabbitMQ 中，默认的 exchange 是一个直连型的 exchange，其名称是空字符串。
	// 当一个队列被声明时，如果没有指定该队列应该绑定的 exchange，则该队列会自动绑定到默认的 exchange 上。
	// 因此，当我们使用 `exchange` 参数为空字符串时，`ch.Publish` 函数会将消息发布到默认的 exchange 上。
	// 默认的 exchange 会将消息按照 routing key 的值，直接路由到名称与该 routing key 相同的队列上。如果该队列不存在，则该消息会被丢弃。
	// 需要注意的是，使用默认的 exchange 进行消息路由时，`routingKey` 参数必须设置为目标队列的名称，否则消息无法正确路由到目标队列。
	// 同时，不同于其他 exchange，使用默认 exchange 发送的消息不能进行多重绑定（multiple bindings），也就是说，每个 routing key 只能与一个队列绑定。
	// 因此，如果需要进行多重绑定，或者自定义路由逻辑，则需要使用其他类型的 exchange。
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		msg)
	if err != nil {
		return fmt.Errorf("%s: %s", "Failed to publish a message", err)
	}
	return nil
}

// Consume 消费端
// queue 队列名称，
// expire 队列过期时间(秒)，用于声明队列时，设置消息的有效期，如果为0则表明永久有效果
// process 处理结果函数（需要在该函数中手动ack或者nack
func (c *Client) Consume(queue string, expire int64, process func(msg amqp.Delivery)) error {
	ch, q, err := c.createChannelAndQueueDeclare(queue, true, expire)
	if err != nil {
		return err
	}
	defer func(ch *amqp.Channel) {
		if err := ch.Close(); err != nil {
			log.Println("close channel err", err.Error())
		}
	}(ch)

	// 如果在 timeout 毫秒内没有调用`Ack()`或`Nack()`方法，消息会自动丢弃，如果配置了死信队列，则丢了死信队列中
	var args amqp.Table
	// if timeout > 0 {
	//	args = amqp.Table{"x-message-ttl": timeout * 1000}
	// }

	// 用于消费消息的方法，用于从指定队列中获取消息并进行处理
	// 第二个参数作为一个消费者标签用于区分多个消费者使用同一队列的情况。如果指定空字符串，则会自动生成一个唯一的。在取消订阅时需要使用该标识。
	// 如果有多个消费者同时从同一个队列中获取消息，建议使用不同的 consumer tag 进行标识，以避免消息重复消费等问题。
	// 注意，如果autoAck为true，则表示收到消息后立即确认并从队列中删除。如果auto-ack为false，则需要手动调用amqp.Deliveries中的Delivery.Ack方法确认收到消息
TRY:
	msgChan, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		args,   // args
	)
	if err != nil {
		return fmt.Errorf("%s: %s", "Failed to publish a message", err)
	}
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				if c.conn.IsClosed() {
					log.Println("Connect closed,try reconnect")
					err := c.reconnect()
					if err != nil {
						log.Println("Try reconnect fail,wait 3 second,", err.Error())
						time.Sleep(3 * time.Second)
					} else {
						log.Println("Try reconnect success")
					}
				} else {
					// 通道关闭
					log.Println("Channel closed,try get open channel")
					ch, q, err = c.createChannelAndQueueDeclare(queue, true, expire)
					if err != nil {
						log.Println("Try reopen channel fail,wait 3 second,", err.Error())
						time.Sleep(3 * time.Second)
					} else {
						log.Println("Try reopen channel success")
						goto TRY
					}
				}
				break
			}
			go process(msg)
		default:

		}
	}
	log.Println("exit")
	return nil
}

func (c *Client) reconnect() error {
	config := amqp.Config{
		Heartbeat: 30 * time.Second,
	}
	var err error
	c.conn, err = amqp.DialConfig(c.opts.Url, config)
	if err != nil {
		return fmt.Errorf("%s: %s", "Failed to reconnect to RabbitMQ", err)
	}

	return nil
}
