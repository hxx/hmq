package model

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var AclColl *mongo.Collection

type Acl struct {
	DeviceSecret string `bson:"device_secret"`
	DeviceID     string `bson:"device_id"`
	ProductID    string `bson:"product_id"`
	PasswordHash string `bson:"password_hash"`
	Role         int    `bson:"role"`
	Username     string `bson:"username"`
	Password     string `bson:"password"`
	CreateTime   string `bson:"create_time"`
	UpdateTime   string `bson:"update_time"`
	TopicList    map[string]string `bson:"topic_list"`
}

//TableName
func (drl Acl) TableName() string {
	return "c_acl"
}

func (drl *Acl) Create() error {
	_, err := AclColl.InsertOne(context.Background(), drl)
	return err
}

func (drl *Acl) One(m bson.M) (acl Acl, err error) {
	err = AclColl.FindOne(context.Background(), m).Decode(&acl)
	return
}

func (drl *Acl) List(m bson.M) (list []Acl, err error) {
	cursor, err := AclColl.Find(context.Background(), m)
	if err != nil {
		return list, err
	}
	for cursor.Next(context.TODO()) {
		var acl Acl
		if err = cursor.Decode(&acl); err != nil {
			return list, err
		}
		list = append(list, acl)
	}
	return list, nil
}
