package result

/*
错误码设计
第一位表示错误级别, 1 为系统错误, 2 为普通错误
第二三位表示服务模块代码
第四五位表示具体错误代码
*/

var (
	SUCCESS = &Errno{200, "success"}

	// 系统错误, 前缀为 100
	InternalServerError = &Errno{10001, "内部服务器错误"}
	ErrBind             = &Errno{10002, "请求参数错误"}
)

func DecodeErr(err error) (int, string) {
	if err == nil {
		return SUCCESS.Code, SUCCESS.Message
	}
	switch typed := err.(type) {
	case *Result:
		if typed.Code == ErrBind.Code {
			typed.Msg = typed.Msg + " 具体是 " + typed.ErrMsg
		}
		return typed.Code, typed.Msg
	case *Errno:
		return typed.Code, typed.Message
	default:
	}

	return InternalServerError.Code, err.Error()
}
