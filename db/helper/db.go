package helper

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
var MongoDB *mongo.Client

type MongoOpt struct {
	Addr     string
	UserName string
	Password string
}

func NewMongo(mo MongoOpt) error {
	// Set Client options
	clientOptions := options.Client().ApplyURI(mo.Addr)
	//clientOptions.Auth = &options.Credential{Username: mo.UserName, Password: mo.Password}
	// Connect to MongoDB
	var err error
	MongoDB, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}
	// Check the connection
	if err = MongoDB.Ping(context.TODO(), nil); err != nil {
		return err
	}
	fmt.Println("mongoDB run success!")

	return nil
}