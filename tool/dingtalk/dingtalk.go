package dingtalk

import (
	"github.com/houyanzu/work-box/config"
	"github.com/houyanzu/work-box/lib/httptool"
)

func Push(msg string) {
	conf := config.GetConfig()
	js := `{"msgtype":"text","text": {"content": "` + msg + `"}}`

	_, _, _ = httptool.PostJSON(conf.Extra.DingTalkURL, []byte(js))
}
