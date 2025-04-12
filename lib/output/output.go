package output

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// ErrorCode 类型用于表示错误码
type ErrorCode int

// Output 结构体用于构建 API 响应
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

// InitMsgMap 初始化错误消息映射
func InitMsgMap(msgMap_ map[ErrorCode]map[string]string) {
	if msgMap_ == nil {
		msgMap = make(map[ErrorCode]map[string]string)
	}
	msgMap = msgMap_
}

// NewOutput 创建一个新的 Output 实例
func NewOutput(c *gin.Context, code ErrorCode, data interface{}) *Output {
	lang := c.GetHeader("Language")
	if lang == "" {
		lang = "zh"
	}

	msg := ""
	if msgs, ok := msgMap[code]; ok {
		msg = msgs[lang]
	}

	return &Output{
		c:     c,
		code:  code,
		data:  data,
		lang:  lang,
		msg:   msg,
		total: -1,
	}
}

// DiyMsg 设置自定义消息
func (out *Output) DiyMsg(msg string) *Output {
	out.msg = msg
	return out
}

// AppendMsg 追加消息
func (out *Output) AppendMsg(msg string) *Output {
	if out.msg == "" {
		out.msg = msg
	} else {
		var builder strings.Builder
		builder.WriteString(out.msg)
		builder.WriteString(msg)
		out.msg = builder.String()
	}
	return out
}

// Total 设置数据总数
func (out *Output) Total(total int64) *Output {
	out.total = total
	return out
}

// GetCode 获取错误码
func (out *Output) GetCode() ErrorCode {
	return out.code
}

// GetMsg 获取消息
func (out *Output) GetMsg() string {
	return out.msg
}

// GetData 获取数据
func (out *Output) GetData() interface{} {
	return out.data
}

// Out 发送 JSON 响应
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
