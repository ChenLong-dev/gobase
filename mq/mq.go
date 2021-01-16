/**
* @Author: cl
* @Date: 2021/1/16 12:09
 */
package mq

import (
	"encoding/json"
	"fmt"
	"github.com/ChenLong-dev/gobase/mlog"
	"github.com/gofrs/uuid"
	"github.com/streadway/amqp"
	"sync"
	"time"
)

// 定义全局变量,指针类型
var mqConn *amqp.Connection
var mqChan *amqp.Channel
// 定义生产者接口
type Producer interface {
	//MsgContent() string
	MsgContent() map[string]interface{}
}

// 定义接收者接口
type Receiver interface {
	Consumer([]byte) error
	//Consumer(map[string]interface{}) error
}

// 定义RabbitMQ对象
type RabbitMQ struct {
	connection   *amqp.Connection
	channel      *amqp.Channel
	taskName     string // 任务名称
	queueName    string // 队列名称
	routingKey   string // key名称
	exchangeName string // 交换机名称
	exchangeType string // 交换机类型
	producerList []Producer
	receiverList []Receiver
	mu           sync.RWMutex
	brokerUrl    string // 连接路径
}

// 定义队列交换机对象
type QueueExchange struct {
	TaskName string // 任务名称
	QuName   string // 队列名称
	RtKey    string // key值
	ExName   string // 交换机名称
	ExType   string // 交换机类型
}

// 链接rabbitMQ
func (r *RabbitMQ) mqConnect() error {
	var err error
	mqConn, err = amqp.Dial(r.brokerUrl)
	r.connection = mqConn // 赋值给RabbitMQ对象
	if err != nil {
		mlog.Errorf("MQ打开链接失败:%s \n", err)
		return err
	}
	mqChan, err = mqConn.Channel()
	r.channel = mqChan // 赋值给RabbitMQ对象
	if err != nil {
		mlog.Errorf("MQ打开管道失败:%s \n", err)
		return err
	}
	//mlog.Info("MQ连接成功 ...")
	return nil
}

// 关闭RabbitMQ连接
func (r *RabbitMQ) mqClose() {
	// 先关闭管道,再关闭链接
	if r.channel != nil {
		err := r.channel.Close()
		if err != nil {
			mlog.Errorf("MQ管道 channel 关闭失败:%s \n", err)
		}
		//mlog.Info("MQ链接 channel 关闭成功 ...")
	}
	if r.connection != nil {
		err := r.connection.Close()
		if err != nil {
			mlog.Errorf("MQ链接 connection 关闭失败:%s \n", err)
		}
		//mlog.Info("MQ链接 connection 关闭成功 ...")
	}

}

// 创建一个新的操作对象
func New(q *QueueExchange, broker string) *RabbitMQ {
	return &RabbitMQ{
		taskName:     q.TaskName,
		queueName:    q.QuName,
		routingKey:   q.RtKey,
		exchangeName: q.ExName,
		exchangeType: q.ExType,
		brokerUrl:    broker,
	}
}

// 启动RabbitMQ客户端,并初始化
func (r *RabbitMQ) Start() {
	// 开启监听生产者发送任务
	for _, producer := range r.producerList {
		go r.listenProducer(producer)
	}
	// 开启监听接收者接收任务
	for _, receiver := range r.receiverList {
		go r.listenReceiver(receiver)
	}
	defer r.mqClose()
	time.Sleep(3 * time.Second)
}

// 注册发送指定队列指定路由的生产者
func (r *RabbitMQ) RegisterProducer(producer Producer) {
	r.producerList = append(r.producerList, producer)
}

type Body2 struct {
}

// celery 特定的消息体
type Body3 struct {
	CallBack *string `json:"callbacks"`
	ErrBacks *string `json:"errbacks"`
	Chain    *string `json:"chain"`
	Chord    *string `json:"chord"`
}

// 定义队列交换机对象
type Headers struct {
	task     string
	id       string
	argsrepr interface{}
}

// 生成mq 消息体
func makeMqMessage(body interface{}) []byte {
	var msg []interface{}
	var body1 []interface{}
	var body2 Body2
	var body3 Body3

	body1 = append(body1, body)

	body3.CallBack = nil
	body3.ErrBacks = nil
	body3.Chain = nil
	body3.Chord = nil

	msg = append(msg, body1)
	msg = append(msg, body2)
	msg = append(msg, body3)
	buf2, _ := json.Marshal(msg)
	//mlog.Debugf("Producer: %s\n", buf2)
	// celery消息模板
	//[[{"key": "val"}], {}, {"callbacks": null, "errbacks": null, "chain": null, "chord": null}]
	return buf2
}

// 发送任务
func (r *RabbitMQ) listenProducer(producer Producer) {
	// 处理结束关闭链接
	defer func() {
		if err := recover(); err != nil {
			mlog.Errorf("[C]MQ连接失败:%s, Recover \n", err)
			return
		}
	}()
	//defer r.mqClose()
	// 验证链接是否正常,否则重新链接
	if r.channel == nil {
		if err := r.mqConnect(); err != nil {
			mlog.Errorf("[P]MQ连接失败:%s \n", err)
			return
		}
	}

	// 用于检查队列是否存在,已经存在不需要重复声明
	_, err := r.channel.QueueDeclarePassive(r.queueName, true, false, false, true, nil)
	if err != nil {
		// 队列不存在,声明队列
		// name:队列名称;durable:是否持久化,队列存盘,true服务重启后信息不会丢失,影响性能;autoDelete:是否自动删除;noWait:是否非阻塞,
		// true为是,不等待RMQ返回信息;args:参数,传nil即可;exclusive:是否设置排他
		_, err = r.channel.QueueDeclare(r.queueName, true, false, false, true, nil)
		if err != nil {
			mlog.Errorf("[P]MQ注册队列失败:%s \n", err)
			return
		}
	}
	// 队列绑定
	err = r.channel.QueueBind(r.queueName, r.routingKey, r.exchangeName, true, nil)
	if err != nil {
		mlog.Errorf("[P]MQ绑定队列失败:%s \n", err)
		return
	}
	// 用于检查交换机是否存在,已经存在不需要重复声明
	err = r.channel.ExchangeDeclarePassive(r.exchangeName, r.exchangeType, true, false, false, true, nil)
	if err != nil {
		// 注册交换机
		// name:交换机名称,kind:交换机类型,durable:是否持久化,队列存盘,true服务重启后信息不会丢失,影响性能;autoDelete:是否自动删除;
		// noWait:是否非阻塞, true为是,不等待RMQ返回信息;args:参数,传nil即可; internal:是否为内部
		err = r.channel.ExchangeDeclare(r.exchangeName, r.exchangeType, true, false, false, true, nil)
		if err != nil {
			mlog.Errorf("[P]MQ注册交换机失败:%s \n", err)
			return
		}
	}

	headers := make(map[string]interface{})
	headers["task"] = r.taskName
	headers["id"] = uuid.Must(uuid.NewV4()).String()
	buf1, _ := json.Marshal(producer.MsgContent())
	headers["argsrepr"] = buf1

	// 发送任务消息
	err = r.channel.Publish(r.exchangeName, r.routingKey, false, false, amqp.Publishing{
		ContentType:     "application/json",
		ContentEncoding: "utf-8",
		Body:            makeMqMessage(producer.MsgContent()),
		DeliveryMode:    amqp.Persistent, //消息是否持久化， 1：不持久化， 2：持久化
		Headers:         headers,
	})
	if err != nil {
		mlog.Errorf("[P]MQ任务发送失败:%s \n", err)
		return
	}
}

// 注册接收指定队列指定路由的数据接收者
func (r *RabbitMQ) RegisterReceiver(receiver Receiver) {
	r.mu.Lock()
	r.receiverList = append(r.receiverList, receiver)
	r.mu.Unlock()
}

// 监听接收者接收任务
func (r *RabbitMQ) listenReceiver(receiver Receiver) {
	// 处理结束关闭链接
	defer func() {
		if err := recover(); err != nil {
			mlog.Errorf("[C]MQ连接失败:%s, Recover \n", err)
			return
		}
	}()
	//defer r.mqClose()
	// 验证链接是否正常
	if r.channel == nil {
		if err := r.mqConnect(); err != nil {
			mlog.Errorf("[C]MQ连接失败:%s \n", err)
			return
		}
	}
	// 用于检查队列是否存在,已经存在不需要重复声明
	_, err := r.channel.QueueDeclarePassive(r.queueName, true, false, false, true, nil)
	if err != nil {
		// 队列不存在,声明队列
		// name:队列名称;durable:是否持久化,队列存盘,true服务重启后信息不会丢失,影响性能;autoDelete:是否自动删除;noWait:是否非阻塞,
		// true为是,不等待RMQ返回信息;args:参数,传nil即可;exclusive:是否设置排他
		_, err = r.channel.QueueDeclare(r.queueName, true, false, false, true, nil)
		if err != nil {
			mlog.Errorf("[C]MQ注册队列失败:%s \n", err)
			return
		}
	}
	// 绑定任务
	err = r.channel.QueueBind(r.queueName, r.routingKey, r.exchangeName, true, nil)
	if err != nil {
		mlog.Errorf("[C]绑定队列失败:%s \n", err)
		return
	}
	// 获取消费通道,确保rabbitMQ一个一个发送消息, 设置每次从消息队列获取任务的数量
	err = r.channel.Qos(
		1,    //预取任务数量
		0,    //预取大小
		true, //全局设置
	)
	if err != nil {
		//无法设置Qos
		fmt.Println(err)
		return
	}

	msgList, err := r.channel.Consume(
		r.queueName, // name
		"",          // consumer
		false,       // autoAck
		false,       // exclusive
		false,       // noLocal
		false,       // noWait
		nil,         // args
	)
	if err != nil {
		mlog.Errorf("[C]获取消费通道异常:%s \n", err)
		return
	}
	for msg := range msgList {
		// 处理数据
		err := receiver.Consumer(msg.Body)
		if err != nil {
			err = msg.Ack(true)
			if err != nil {
				mlog.Errorf("[C]确认消息未完成异常:%s \n", err)
				return
			}
		} else {
			// 确认消息,必须为false
			err = msg.Ack(false)
			if err != nil {
				mlog.Errorf("[C]确认消息完成异常:%s \n", err)
				return
			}
			return
		}
	}
}

func GetData(dataByte []byte) (map[string]interface{}, error) {
	//mlog.Infof("Consumer-1:%T, %T, %s\n", string(dataByte), dataByte, string(dataByte))
	var dataArray []interface{}
	err := json.Unmarshal(dataByte, &dataArray)
	if err != nil {
		mlog.Error(err)
		return nil, err
	}

	mapList := dataArray[0].([]interface{})
	return mapList[0].(map[string]interface{}), nil
}

//type TestPro struct {
//	msgContent   map[string]interface{}
//}
//
//// 实现发送者
//func (t *TestPro) MsgContent() map[string]interface{} {
//	return t.msgContent
//}
//
//// 实现接收者
//func (t *TestPro) Consumer(dataByte []byte) error {
//	fmt.Println("-----Consumer:", string(dataByte))
//	return nil
//}
//
//func testMq() {
//	body := make(map[string]interface{})
//
//
//	//指定发送队列与任务名
//	queueExchange := &mq.QueueExchange{
//		"test.task.name",
//		"cl_config_manage_queue",
//		"cl_config_manage_queue",
//		"cl_config_manage_exchange",
//		"direct",
//	}
//	// 发送消息到mq
//	mq := mq.New(queueExchange, "amqp://root:sangfor123@10.227.63.170:5672/")
//	for i := 0; i < 5; i++ {
//		body["key"+ strconv.Itoa(i)] = "hellov + " + strconv.Itoa(i)
//	}
//	t := &TestPro{ body, }
//	mq.RegisterProducer(t)
//	mq.Start()
//
//	time.Sleep(time.Second * 10)
//
//	mq.RegisterReceiver(t)
//	mq.Start()
//
//	time.Sleep(time.Second * 10)
//
//	mq.RegisterReceiver(t)
//	mq.Start()
//
//}