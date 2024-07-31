package dingtalk

import (
	"errors"
	"github.com/houyanzu/work-box/config"
	"github.com/houyanzu/work-box/lib/httptool"
)

func Push(msg string) {
	conf := config.GetConfig()
	js := `{"msgtype":"text","text": {"content": "` + msg + `"}}`
	if conf.Extra.DingTalkURL == "" {
		return
	}

	_, _, _ = httptool.PostJSON(conf.Extra.DingTalkURL, []byte(js))
}

func PushMsg(msg string) error {
	conf := config.GetConfig()
	js := `{"msgtype":"text","text": {"content": "` + msg + `"}}`
	if conf.Extra.DingTalkURL == "" {
		return errors.New("empty dingtalk url")
	}

	_, _, _ = httptool.PostJSON(conf.Extra.DingTalkURL, []byte(js))
	return nil
}
