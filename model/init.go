package model

import (
	"fmt"
	_ "github.com/fhmq/hmq/config"
	"github.com/fhmq/hmq/db/helper"
	"github.com/spf13/viper"
)

func init() {
	var mongoOpt = helper.MongoOpt{
		Addr:     viper.GetString("mongo.addr"),
		UserName: viper.GetString("mongo.username"),
		Password: viper.GetString("mongo.password"),
	}

	//初始化mongo
	if err := helper.NewMongo(mongoOpt); err != nil {
		panic(fmt.Sprintf("初始化mongo客户端 err: %v", err))
	}

	database := viper.GetString("mongo.database")

	//初始化表格
	AclColl = helper.MongoDB.Database(database).Collection(Acl{}.TableName())
}
