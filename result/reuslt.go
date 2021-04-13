package result

import (
	"fmt"
	"net/http"

	"github.com/fhmq/hmq/loge"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 定义错误码
type Errno struct {
	Code    int
	Message string
}

func (err Errno) Error() string {
	return err.Message
}

// 定义错误
type Result struct {
	RequestId interface{} `json:"request_id"` //服务请求request_id
	Code      int         `json:"code"`       // 错误码
	Msg       string      `json:"msg"`        // 展示给用户看的
	ErrMsg    string      `json:"err_msg"`    // 保存内部错误信息
	Data      interface{} `json:"data"`       //返回数据
}

func (err *Result) Error() string {
	return fmt.Sprintf("Result - code: %d, message: %s, error: %s", err.Code, err.Msg, err.ErrMsg)
}

func (r *Result) E(requestId interface{}) *Result {
	loge.ResultError(zap.String("E", "E"), r.Error(), zap.Any("requestId", requestId))
	return r
}

func New(c *gin.Context, e *Result) {
	requestId := c.MustGet("request_id")
	api := fmt.Sprintf("%s:%s", c.Request.Method, c.Request.URL)
	e.RequestId = requestId
	loge.ResultError(zap.String("api", api), zap.Any("err", e.ErrMsg), zap.Any("errMsg", e.Msg), zap.Any("requestId", requestId))
	c.AbortWithStatusJSON(http.StatusOK, e)
}

func (errno *Errno) ErrMsg(err error) *Result {
	errord := ""
	if err != nil {
		errord = err.Error()
	}
	return &Result{
		Code:   errno.Code,
		Msg:    errno.Message,
		ErrMsg: errord,
		Data:   map[string]interface{}{},
	}
}

func (errno *Errno) ErrMsgStr(errStr string) *Result {
	return &Result{
		Code:   errno.Code,
		Msg:    errno.Message,
		ErrMsg: errStr,
		Data:   map[string]interface{}{},
	}
}

func (e *Errno) Msg() *Result {
	return &Result{
		Code:   e.Code,
		Msg:    e.Message,
		ErrMsg: "",
		Data:   map[string]interface{}{},
	}
}

func (e *Errno) Ok(data interface{}) *Result {
	return &Result{
		Code:   e.Code,
		Msg:    e.Message,
		ErrMsg: "",
		Data:   data,
	}
}

func Ok(c *gin.Context, data interface{}) {
	requestId := c.MustGet("request_id")
	if data == nil {
		data = make(map[string]interface{})
	}
	c.AbortWithStatusJSON(200, Result{
		Code:      SUCCESS.Code,
		Msg:       SUCCESS.Message,
		ErrMsg:    "",
		RequestId: requestId,
		Data:      data,
	})
}

func ServerError(c *gin.Context, errMsg error) {
	api := fmt.Sprintf("%s:%s", c.Request.Method, c.Request.URL)
	requestId := c.MustGet("request_id")
	loge.ResultError(zap.String("api", api), zap.Error(errMsg), zap.String("errMsg", InternalServerError.Message), zap.Any("requestId", requestId.(string)))
	data := make(map[string]interface{})
	c.AbortWithStatusJSON(200, Result{
		Code:      InternalServerError.Code,
		Msg:       InternalServerError.Message,
		ErrMsg:    fmt.Sprintf("%v", errMsg),
		Data:      data,
		RequestId: requestId,
	})
}
