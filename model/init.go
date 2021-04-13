package model

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
)

type Config struct {
	Addr  string `json:"addr"`
	Database   string `json:"database"`
	Password string `json:"password"`
	Username string `json:"username"`
}

var (
	config Config
)

func init() {
	content, err := ioutil.ReadFile("./plugins/auth/authmango/mango.json")
	if err != nil {
		log.Fatal("Read config file error: ", zap.Error(err))
	}
	// log.Info(string(content))

	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Unmarshal config file error: ", zap.Error(err))
	}
	// fmt.Println("http: config: ", config)

	clientOptions := options.Client().ApplyURI(config.Addr)
	//clientOptions.Auth = &options.Credential{Username: mo.UserName, Password: mo.Password}
	// Connect to MongoDB
	mongoDB, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("mongo connect failed: ", zap.Error(err))
	}
	// Check the connection
	if err = mongoDB.Ping(context.TODO(), nil); err != nil {
		log.Fatal("mongo ping failed: ", zap.Error(err))
	}
	fmt.Println("mongoDB run success!")

	database := config.Database

	//初始化表格
	AclColl = mongoDB.Database(database).Collection(Acl{}.TableName())
	LicenseBatchColl = mongoDB.Database(database).Collection(LicenseBatch{}.TableName())
}
