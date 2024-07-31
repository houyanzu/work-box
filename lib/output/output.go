package output

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type ErrorCode int

type Output struct {
	c *gin.Context

	code  ErrorCode
	msg   string
	data  interface{}
	lang  string
	other interface{}
	total int64
}

var msgMap map[ErrorCode]map[string]string

func InitMsgMap(msgMap_ map[ErrorCode]map[string]string) {
	msgMap = msgMap_
}

func NewOutput(c *gin.Context, code ErrorCode, data interface{}) *Output {
	var out Output
	out.c = c
	out.data = data

	out.lang = c.GetHeader("Language")
	if out.lang == "" {
		out.lang = "zh"
	}

	out.total = -1
	out.msg = msgMap[code][out.lang]
	out.code = code
	return &out
}

func (out *Output) DiyMsg(msg string) *Output {
	out.msg = msg
	return out
}

func (out *Output) AppendMsg(msg string) *Output {
	out.msg = out.msg + msg
	return out
}

func (out *Output) Total(total int64) *Output {
	out.total = total
	return out
}

func (out *Output) GetCode() ErrorCode {
	return out.code
}

func (out *Output) GetMsg() string {
	return out.msg
}

func (out *Output) GetData() interface{} {
	return out.data
}

func (out *Output) Out() {
	if out.data == nil {
		if out.total == 0 {
			out.data = make([]interface{}, 0)
		} else {
			out.data = gin.H{}
		}
	}
	result := gin.H{
		"code": out.code,
		"msg":  out.msg,
		"data": out.data,
	}
	if out.total >= 0 {
		result["total"] = out.total
	}
	out.c.JSON(http.StatusOK, result)
}
