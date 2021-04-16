package bridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/nats-io/stan.go"
	"go.uber.org/zap"
)

type natsConfig struct {
	Addr       string            `json:"addr"`
	ClusterId  string            `json:"cluster_id"`
	ClientId   string            `json:"client_id"`
	StateTopic string            `json:"state_topic"`
	DeliverMap map[string]string `json:"deliver_map"`
}

type natsClient struct {
	natsConfig natsConfig
	natsConn   stan.Conn
}

type EventMsg struct {
	ClientID  string `json:"client_id"`
	ClientIP  string `json:"client_ip"`
	Ts        int64  `json:"ts"`
	Payload   string `json:"payload,omitempty"`
	EventType string `json:"event_type"`
}

//Init init nats client
func InitNats() *natsClient {
	log.Info("start connect nats....")
	content, err := ioutil.ReadFile("./plugins/bridge/nats/nats.json")
	if err != nil {
		log.Fatal("Read config file error: ", zap.Error(err))
	}
	// log.Info(string(content))
	var config natsConfig
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Unmarshal config file error: ", zap.Error(err))
	}
	c := &natsClient{natsConfig: config}
	c.connect()
	return c
}

//connect
func (n *natsClient) connect() {
	sc, err := stan.Connect(n.natsConfig.ClusterId, fmt.Sprintf("hmq-%s", n.natsConfig.ClientId), stan.NatsURL(n.natsConfig.Addr))
	if err != nil {
		log.Fatal("create nats async producer failed: ", zap.Error(err))
	}

	n.natsConn = sc
}

//Publish publish to nats
func (n *natsClient) Publish(e *Elements) error {
	log.Debug("element: ", zap.Any("element", e))
	switch e.Action {
	case Connect, Disconnect:
		eMsg := &EventMsg{
			ClientID:  e.ClientID,
			ClientIP:  e.ClientIP,
			Ts:        e.Timestamp,
			EventType: e.Action,
		}
		eMsgByte, err := json.Marshal(eMsg)
		if err != nil {
			return err
		}
		return n.natsConn.Publish(n.natsConfig.StateTopic, eMsgByte)
	case Publish:
		if e.Topic != "" {
			for reg, topic := range n.natsConfig.DeliverMap {
				match := matchTopic(reg, e.Topic)
				if match {
					eMsg := &EventMsg{
						ClientID:  e.ClientID,
						ClientIP:  e.ClientIP,
						Ts:        e.Timestamp,
						Payload:   e.Payload,
						EventType: e.Action,
					}
					eMsgByte, err := json.Marshal(eMsg)
					if err != nil {
						return err
					}
					return n.natsConn.Publish(topic, eMsgByte)
				}
			}
		}
	default:
		return errors.New("error action: " + e.Action)
	}
	return nil
}
