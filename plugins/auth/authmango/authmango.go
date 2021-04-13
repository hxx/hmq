package authmango

import (
	"github.com/fhmq/hmq/logger"
	"github.com/fhmq/hmq/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"strings"
)

type authMango struct{}

var (
	log = logger.Get().Named("authmango")
)

const (
	SUB      = "1"
	PUB      = "2"
	PUBSUB   = "3"
)

//Init init authMango
func Init() *authMango {
	return &authMango{}
}

//CheckAuth check mqtt connect
func (a *authMango) CheckConnect(clientID, username, password string) bool {
	action := "connect"
	{
		aCache := checkCache(action, clientID, username, password, "")
		if aCache != nil {
			if aCache.password == password && aCache.username == username && aCache.action == action && aCache.clientID == clientID {
				return true
			}
		}
	}

	/*
		ClientID {ProductKey}&{DeviceName}
		Username {ProductKey}&{DeviceID}
		Password 算法获取
	*/

	productId := strings.Split(clientID, "&")[0]
	deviceId := strings.Split(clientID, "&")[1]

	if len(productId) == 0 {
		log.Error("productId is empty: ")
		return false
	}

	if len(deviceId) == 0 {
		log.Error("deviceId is empty: ")
		return false
	}

	var acl model.Acl
	m := bson.M{
		"product_id": productId,
		"device_id":  deviceId,
		"username":   username,
		"password":   password,
	}
	acl, err := acl.One(m)
	if err != nil {
		log.Error("can't found acl in mongo: ", zap.Error(err))
		return false
	}
	addCache(action, clientID, username, password, "")
	return true
}

//CheckACL check mqtt connect
func (a *authMango) CheckACL(action, clientID, username, ip, topic string) bool {

	{
		aCache := checkCache(action, clientID, "", "", topic)
		if aCache != nil {
			if aCache.topic == topic && aCache.action == action && aCache.clientID == clientID {
				return true
			}
		}
	}

	sysLabel := strings.Split(topic, "/")[0]

	if len(sysLabel) < 0 {
		log.Error("no sysLabel")
		return false
	}

	if sysLabel == "$rrpc" || sysLabel == "$data" {
		addCache(action, clientID, "", "", topic)
		return true
	}

	// 自定义 topic
	productId := strings.Split(clientID, "&")[0]
	deviceId := strings.Split(clientID, "&")[1]

	if len(productId) == 0 {
		log.Error("productId is empty: ")
		return false
	}

	if len(deviceId) == 0 {
		log.Error("deviceId is empty: ")
		return false
	}

	var acl model.Acl
	m := bson.M{
		"product_id": productId,
		"device_id":  deviceId,
		"username":   username,
	}
	acl, err := acl.One(m)
	if err != nil {
		log.Error("can't found acl in mongo: ", zap.Error(err))
		return false
	}

	if _, ok := acl.TopicList[topic]; !ok {
		log.Error("custom topic not in topic list.")
		return false
	}

	// 发布动作时，检测该 topic 是否是只读
	if action == PUB || action == PUBSUB {
		if acl.TopicList[topic] == "r" {
			log.Error("this custom topic can't publish")
			return false
		}
	}

	addCache(action, clientID, "", "", topic)
	return true
}
