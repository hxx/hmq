package broker

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fhmq/hmq/loge"
	"github.com/fhmq/hmq/result"

	"github.com/fhmq/hmq/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

var router = gin.Default()
var RootPath = ""

type licensePostReq struct {
	ProductId string `json:"product_id" binding:"required"`
	Amount    int    `json:"amount" binding:"required"`
}

type licenseInfo struct {
	ProductId string `json:"product_id"`
	Amount    int    `json:"amount"`
	BatchId   string `json:"batch_id"`
}

type aclInfo struct {
	ProductId    string `json:"product_id"`
	DeviceId     string `json:"device_id"`
	DeviceSecret string `json:"device_secret"`
}

func GetProjectPath() string {
	var projectPath string
	projectPath, _ = os.Getwd()
	return projectPath
}

func InitHTTPMoniter(b *Broker) {
	RootPath = os.Getenv("RootPath")
	if RootPath == "" {
		RootPath = filepath.Join(GetProjectPath(), "logs")
	}
	loge.New(RootPath, "http_server")

	//request_id生成
	router.Use(func(context *gin.Context) {
		context.Set("request_id", strings.ReplaceAll(uuid.NewV4().String(), "-", ""))
		context.Next()
	})

	// 检查未完成的 license batch
	go initCheckAcl()

	serverRun(b)
}

func serverRun(b *Broker) {
	//------------------健康检查接口
	router.GET("/health", func(context *gin.Context) {
		loge.Info("Health")
		context.JSON(200, "ok")
	})

	router.DELETE("api/v1/connections/:clientid", func(c *gin.Context) {
		clientid := c.Param("clientid")
		cli, ok := b.clients.Load(clientid)
		if ok {
			conn, succss := cli.(*client)
			if succss {
				conn.Close()
			}
		}
		resp := map[string]int{
			"code": 0,
		}
		c.JSON(200, &resp)
	})

	// license申请
	router.POST("api/v1/licenses", func(c *gin.Context) {
		requestId := c.MustGet("request_id").(string)
		var req licensePostReq
		if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
			loge.Error(zap.Any("bind body with json failed", err), zap.Any("request_id", requestId))
			result.New(c, result.ErrBind.ErrMsg(err))
			return
		}

		if req.Amount > 2000 {
			loge.Error(zap.Any("amount cannot be greater than 2000", req.Amount), zap.Any("request_id", requestId))
			result.New(c, result.ErrBind.ErrMsg(fmt.Errorf("amount cannot be greater than 2000")))
			return
		}

		productId := req.ProductId
		amount := req.Amount
		batchId := strings.ReplaceAll(uuid.NewV4().String(), "-", "")
		currentTime := time.Now().Format("2006/01/02 15:04:05")

		// 创建 license batch
		lb := &model.LicenseBatch{
			BatchId:    batchId,
			ProductId:  productId,
			Amount:     amount,
			Status:     1,
			CreateTime: currentTime,
			UpdateTime: currentTime,
		}
		err := lb.Create()
		if err != nil {
			loge.Error(zap.Any("create license batch failed.", err), zap.Any("lb", lb), zap.Any("request_id", requestId))
			result.New(c, result.InternalServerError.ErrMsg(err))
			return
		}

		// 创建 acl
		go checkAndCreateAcl(*lb, requestId)

		li := licenseInfo{
			BatchId:   batchId,
			ProductId: productId,
			Amount:    amount,
		}

		result.Ok(c, li)
	})

	// license查询
	router.GET("api/v1/licenses/:batch_id", func(c *gin.Context) {
		requestId := c.MustGet("request_id").(string)
		batchId := c.Param("batch_id")
		var lb model.LicenseBatch
		lbm := bson.M{
			"batch_id": batchId,
		}
		lb, err := lb.One(lbm)

		if err != nil {
			loge.Error(zap.Any("get license batch failed", err), zap.Any("request_id", requestId))
			result.New(c, result.InternalServerError.ErrMsg(err))
			return
		}

		var acl model.Acl
		m := bson.M{
			"batch_id": batchId,
		}
		aclList, err := acl.List(m)

		if err != nil {
			loge.Error(zap.Any("get acl failed", err), zap.Any("request_id", requestId))
			result.New(c, result.InternalServerError.ErrMsg(err))
			return
		}

		var acls []aclInfo
		for _, a := range aclList {
			acl := aclInfo{
				ProductId:    a.ProductID,
				DeviceId:     a.DeviceID,
				DeviceSecret: a.DeviceSecret,
			}
			acls = append(acls, acl)
		}

		result.Ok(c, map[string]interface{}{
			"batch_id": lb.BatchId,
			"amount":   lb.Amount,
			"status":   lb.Status,
			"list":     acls,
		})
	})

	router.Run(":" + b.config.HTTPPort)
}

func createAcl(batchId, productId, requestId string) {
	deviceId := strings.ReplaceAll(uuid.NewV4().String(), "-", "")
	deviceSecret := strings.ReplaceAll(uuid.NewV4().String(), "-", "")[0:18]
	currentTime := time.Now().Format("2006/01/02 15:04:05")
	username := fmt.Sprintf("%s&%s", productId, deviceId)

	// 默认sha1
	mac := hmac.New(sha1.New, []byte(deviceSecret))
	mac.Write([]byte(fmt.Sprintf("%s%s", productId, deviceId)))
	password := fmt.Sprintf("%x", mac.Sum(nil))

	acl := &model.Acl{
		BatchId:      batchId,
		ProductID:    productId,
		DeviceSecret: deviceSecret,
		DeviceID:     deviceId,
		PasswordHash: "sha1",
		Role:         3, // 普通设备
		Username:     username,
		Password:     password,
		CreateTime:   currentTime,
		UpdateTime:   currentTime,
		TopicList:    map[string]string{},
	}
	err := acl.Create()
	if err != nil {
		loge.Error(zap.Any("create acl failed.", err), zap.Any("acl", acl), zap.Any("request_id", requestId))
	}
	loge.Debug(zap.Any("acl", *acl), zap.Any("request_id", requestId))
}

func checkAndCreateAcl(lb model.LicenseBatch, requestId string) {
	// 检测创建的acl数量
	var acl model.Acl
	m := bson.M{
		"batch_id": lb.BatchId,
	}
	aclList, err := acl.List(m)

	if err != nil {
		loge.Error(zap.Any("get acl failed", err), zap.Any("request_id", requestId))
		return
	}
	loge.Debug(zap.Any("license batch", lb), zap.Any("aclList count", len(aclList)), zap.Any("aclList", aclList), zap.Any("request_id", requestId))

	remaining := lb.Amount-len(aclList)
	for i := 0; i < remaining; i++ {
		createAcl(lb.BatchId, lb.ProductId, requestId)
	}

	// 已创建数量达到目标数量时，改变 license batch 状态
	if len(aclList) == lb.Amount {
		filter := bson.M{"batch_id": bson.M{"$eq": lb.BatchId}}
		update := bson.M{"$set": bson.M{"status": 2}}
		if err := lb.UpdateOne(filter, update); err != nil {
			loge.Error(zap.Any("update license batch failed.", err), zap.Any("lb", lb), zap.Any("request_id", requestId))
		}
		return
	}

	// 剩余数量大于0时，继续创建acl
	if remaining > 0 {
		checkAndCreateAcl(lb, requestId)
		return
	}

	return
}

func initCheckAcl() {
	requestId := fmt.Sprintf("%s_%s", "init_request_id", strings.ReplaceAll(uuid.NewV4().String(), "-", ""))
	var lb model.LicenseBatch
	m := bson.M{}
	lbList, err := lb.List(m)
	if err != nil {
		loge.Error(zap.Any("get license batch list failed", err), zap.Any("request_id", requestId))
		return
	}
	loge.Debug(zap.Any("lbList", lbList), zap.Any("request_id", requestId))

	for _, lb := range lbList {
		if lb.Status == 1 {
			checkAndCreateAcl(lb, requestId)
		}
	}
	return
}