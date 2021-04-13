package model

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var LicenseBatchColl *mongo.Collection

type LicenseBatch struct {
	BatchId    string `bson:"batch_id"`
	ProductId  string `bson:"product_id"`
	Amount     int    `bson:"amount"`
	Status     int    `bson:"status"`
	CreateTime string `bson:"create_time"`
	UpdateTime string `bson:"update_time"`
}

//TableName
func (drl LicenseBatch) TableName() string {
	return "c_license_batch"
}

func (drl *LicenseBatch) Create() error {
	_, err := LicenseBatchColl.InsertOne(context.Background(), drl)
	return err
}

func (drl *LicenseBatch) One(m bson.M) (lb LicenseBatch, err error) {
	err = LicenseBatchColl.FindOne(context.Background(), m).Decode(&lb)
	return
}

func (drl *LicenseBatch) UpdateOne(m bson.M, u bson.M) error {
	_, err := LicenseBatchColl.UpdateOne(context.Background(), m, u)
	return err
}

func (drl *LicenseBatch) List(m bson.M) (list []LicenseBatch, err error) {
	cursor, err := LicenseBatchColl.Find(context.Background(), m)
	if err != nil {
		return list, err
	}
	for cursor.Next(context.TODO()) {
		var lb LicenseBatch
		if err = cursor.Decode(&lb); err != nil {
			return list, err
		}
		list = append(list, lb)
	}
	return list, nil
}
