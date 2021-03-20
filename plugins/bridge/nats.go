package bridge

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/nats-io/stan.go"
	"go.uber.org/zap"
)

type natsConfig struct {
	Addr      string `json:"addr"`
	ClusterId string `json:"cluster_id"`
	ClientId  string `json:"client_id"`
	Topic     string `json:"topic"`
}

type natsClient struct {
	natsConfig natsConfig
	natsConn   stan.Conn
}

//Init init nats client
func InitNats() *natsClient {
	log.Info("start connect nats....")
	content, err := ioutil.ReadFile("./plugins/bridge/nats/nats.json")
	if err != nil {
		log.Fatal(
			"Read config file error: ",
			zap.Error(err),
		)
	}
	// log.Info(string(content))
	var config natsConfig
	err = json.Unmarshal(
		content,
		&config,
	)
	if err != nil {
		log.Fatal(
			"Unmarshal config file error: ",
			zap.Error(err),
		)
	}
	c := &natsClient{natsConfig: config}
	c.connect()
	return c
}

//connect
func (n *natsClient) connect() {
	sc, err := stan.Connect(n.natsConfig.ClusterId, fmt.Sprintf("hmq-%s", n.natsConfig.ClientId), stan.NatsURL(n.natsConfig.Addr))
	if err != nil {
		log.Fatal(
			"create nats async producer failed: ",
			zap.Error(err),
		)
	}

	n.natsConn = sc
}

//Publish publish to nats
func (n *natsClient) Publish(e *Elements) error {
	log.Debug(
		"element: ",
		zap.Any(
			"element",
			e,
		),
	)
	switch e.Action {
	case Publish:
		if e.Topic != "" {
			subj := e.Topic
			if err := n.publish(
				subj,
				e.ClientID,
				e.Payload,
			); err != nil {
				return err
			}
		}
	default:
		return nil
	}
	return nil
}

func (n *natsClient) publish(topic string, clientId string, msg string) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err := n.natsConn.Publish(
		topic,
		[]byte(fmt.Sprintf(
			"%s: %s",
			clientId,
			payload,
		)),
	); err != nil {
		return err
	}

	return nil
}
