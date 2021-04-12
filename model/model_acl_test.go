package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"testing"

	"gotest.tools/assert"
)

func TestAclCreate(t *testing.T) {
	acl := &Acl{
		DeviceSecret: "xxxxxxxxxxxxxxxxxxxxx",
		DeviceID:     "xxxxxxxxxxxxxxxxxxxxx",
		ProductID:    "xxxxxx",
		PasswordHash: "sha1",
		Role:         1,
		Username:     "hxx",
		Password:     "123",
		CreateTime:   "2020-02-20 12:12:14",
		UpdateTime:   "2020-02-20 12:12:14",
		TopicList: map[string]string{
			"xxx/xxxx/control": "r",
			"xxx/xxxx/data":    "rw",
		},
	}
	err := acl.Create()
	assert.Equal(t, err, nil)
}

func TestAclOne(t *testing.T) {
	var acl Acl
	m := bson.M{
		"username":   "hxx",
		"password":   "123",
		"product_id": "xxxxxx",
		"device_id": "xxxxxxxxxxxxxxxxxxxxx",
	}
	acl, err := acl.One(m)
	fmt.Printf("acl: %v", acl)
	assert.Equal(t, err, nil)
}

func TestAclList(t *testing.T) {
	var acl Acl
	m := bson.M{
		"username":   "hxx",
		"password":   "123",
		"product_id": "xxxxxx",
		"device_id": "xxxxxxxxxxxxxxxxxxxxx",
	}
	aclList, err := acl.List(m)
	fmt.Printf("aclList: %v", aclList)
	assert.Equal(t, err, nil)
}
