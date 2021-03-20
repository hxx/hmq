package bridge

import (
	"github.com/fhmq/hmq/logger"
	"strings"
)

const (
	//Connect mqtt connect
	Connect = "connect"
	//Publish mqtt publish
	Publish = "publish"
	//Subscribe mqtt sub
	Subscribe = "subscribe"
	//Unsubscribe mqtt sub
	Unsubscribe = "unsubscribe"
	//Disconnect mqtt disconenct
	Disconnect = "disconnect"
)

var (
	log = logger.Get().Named("bridge")
)

//Elements kafka publish elements
type Elements struct {
	ClientID  string `json:"clientid"`
	Username  string `json:"username"`
	Topic     string `json:"topic"`
	Payload   string `json:"payload"`
	Timestamp int64  `json:"ts"`
	Size      int32  `json:"size"`
	Action    string `json:"action"`
}

const (
	//Kafka plugin name
	Kafka = "kafka"
	//Nats plugin name
	Nats = "nats"
)

type BridgeMQ interface {
	Publish(e *Elements) error
}

func NewBridgeMQ(name string) BridgeMQ {
	switch name {
	case Kafka:
		return InitKafka()
	case Nats:
		return InitNats()
	default:
		return &mockMQ{}
	}
}

func match(subTopic []string, topic []string) bool {
	if len(subTopic) == 0 {
		if len(topic) == 0 {
			return true
		}
		return false
	}

	if len(topic) == 0 {
		if subTopic[0] == "#" {
			return true
		}
		return false
	}

	if subTopic[0] == "#" {
		return true
	}

	// {product_id}/{device_id}/data
	if topic[0] == "data" {
		return true
	}

	if (subTopic[0] == "+") || (subTopic[0] == topic[0]) {
		return match(subTopic[1:], topic[1:])
	}
	return false
}

func matchTopic(subTopic string, topic string) bool {
	return match(strings.Split(subTopic, "/"), strings.Split(topic, "/"))
}