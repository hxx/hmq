package bridge

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"time"

	"github.com/Shopify/sarama"
	"go.uber.org/zap"
)

type kafakConfig struct {
	Addr             []string          `json:"addr"`
	ConnectTopic     string            `json:"onConnect"`
	SubscribeTopic   string            `json:"onSubscribe"`
	PublishTopic     string            `json:"onPublish"`
	UnsubscribeTopic string            `json:"onUnsubscribe"`
	DisconnectTopic  string            `json:"onDisconnect"`
	DeliverMap       map[string]string `json:"deliverMap"`
}

type kafka struct {
	kafakConfig kafakConfig
	kafkaClient sarama.AsyncProducer
}

//Init init kafak client
func InitKafka() *kafka {
	log.Info("start connect kafka....")
	content, err := ioutil.ReadFile("./plugins/bridge/kafka/kafka.json")
	if err != nil {
		log.Fatal("Read config file error: ", zap.Error(err))
	}
	// log.Info(string(content))
	var config kafakConfig
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Unmarshal config file error: ", zap.Error(err))
	}
	c := &kafka{kafakConfig: config}
	c.connect()
	return c
}

//connect
func (k *kafka) connect() {
	conf := sarama.NewConfig()
	conf.Version = sarama.V1_1_1_0
	kafkaClient, err := sarama.NewAsyncProducer(k.kafakConfig.Addr, conf)
	if err != nil {
		log.Fatal("create kafka async producer failed: ", zap.Error(err))
	}

	go func() {
		for err := range kafkaClient.Errors() {
			log.Error("send msg to kafka failed: ", zap.Error(err))
		}
	}()

	k.kafkaClient = kafkaClient
}

//Publish publish to kafka
func (k *kafka) Publish(e *Elements) error {
	config := k.kafakConfig
	key := e.ClientID
	topics := make(map[string]bool)
	switch e.Action {
	case Connect:
		if config.ConnectTopic != "" {
			topics[config.ConnectTopic] = true
		}
	case Publish:
		if config.PublishTopic != "" {
			topics[config.PublishTopic] = true
		}
		// foreach regexp map config
		for reg, topic := range config.DeliverMap {
			match := matchTopic(reg, e.Topic)
			if match {
				topics[topic] = true
			}
		}
	case Subscribe:
		if config.SubscribeTopic != "" {
			topics[config.SubscribeTopic] = true
		}
	case Unsubscribe:
		if config.UnsubscribeTopic != "" {
			topics[config.UnsubscribeTopic] = true
		}
	case Disconnect:
		if config.DisconnectTopic != "" {
			topics[config.DisconnectTopic] = true
		}
	default:
		return errors.New("error action: " + e.Action)
	}

	return k.publish(topics, key, e)

}

func (k *kafka) publish(topics map[string]bool, key string, msg *Elements) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	for topic, _ := range topics {
		select {
		case k.kafkaClient.Input() <- &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.ByteEncoder(key),
			Value: sarama.ByteEncoder(payload),
		}:
			continue
		case <-time.After(5 * time.Second):
			return errors.New("write kafka timeout")
		}

	}

	return nil
}
