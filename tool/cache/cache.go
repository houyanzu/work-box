package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/houyanzu/work-box/config"
	"strconv"
	"time"

	imcache "github.com/houyanzu/cache"
	_ "github.com/houyanzu/cache/redis"
)

var ca imcache.Cache
var prefix string

// InitCache .
func InitCache() error {
	conf := config.GetConfig()
	conn := `"redis://` + conf.Redis.Password + `@` + conf.Redis.Host + `:` + conf.Redis.Port + `"`
	if conf.Redis.User != "" {
		conn = `"redis://` + conf.Redis.User + `:` + conf.Redis.Password + `@` + conf.Redis.Host + `:` + conf.Redis.Port + `"`
	}
	redisConfig := `{
		"conn": ` + conn + `,
		"dbNum": "` + fmt.Sprintf("%v", conf.Redis.Db) + `",
		"key": ""
	}`
	//fmt.Println(redisConfig)
	var err error
	ca, err = imcache.NewCache("redis", redisConfig)
	if err != nil {
		return err
	}
	prefix = conf.Redis.Prefix
	return nil
}

// Set .
func Set(key string, value interface{}, timeout int64) error {
	if prefix != "" {
		key = prefix + key
	}
	if ca == nil {
		panic("cache 未初始化")
	}
	//fmt.Printf("%+v", ca)
	return ca.Put(key, value, time.Duration(timeout)*time.Second)
}

func SetNX(key string, timeout int64) (bool, error) {
	if prefix != "" {
		key = prefix + key
	}
	if ca == nil {
		panic("cache 未初始化")
	}
	return ca.SetNX(key, time.Duration(timeout)*time.Second)
}

// Get .
func Get(key string) interface{} {
	if prefix != "" {
		key = prefix + key
	}
	if ca == nil {
		panic("cache 未初始化")
	}
	return ca.Get(key)
}

func GetInt64(key string) (int64, error) {
	if prefix != "" {
		key = prefix + key
	}
	if ca == nil {
		panic("cache 未初始化")
	}
	resByte, ok := ca.Get(key).([]byte)
	if !ok {
		return 0, errors.New("wrong")
	}
	res, err := strconv.ParseInt(string(resByte), 10, 64)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func GetString(key string) (string, error) {
	if prefix != "" {
		key = prefix + key
	}
	if ca == nil {
		panic("cache 未初始化")
	}
	resByte, ok := ca.Get(key).([]byte)
	if !ok {
		return "", errors.New("wrong")
	}
	return string(resByte), nil
}

func GetByte(key string) (res []byte, err error) {
	if prefix != "" {
		key = prefix + key
	}
	if ca == nil {
		panic("cache 未初始化")
	}
	res, ok := ca.Get(key).([]byte)
	if !ok {
		return nil, errors.New("wrong")
	}
	return
}

func GetObj[T any](key string) (res T, err error) {
	if prefix != "" {
		key = prefix + key
	}

	resByte, err := GetByte(key)
	if err != nil {
		return
	}
	err = json.Unmarshal(resByte, &res)
	return
}

func SetObj[T any](key string, value T, timeout int64) (err error) {
	if prefix != "" {
		key = prefix + key
	}

	js, err := json.Marshal(&value)
	if err != nil {
		return
	}
	err = Set(key, js, timeout)
	return
}

func GetUint64(key string) (uint64, error) {
	if prefix != "" {
		key = prefix + key
	}
	if ca == nil {
		panic("cache 未初始化")
	}
	resByte, ok := ca.Get(key).([]byte)
	if !ok {
		return 0, errors.New("wrong")
	}
	res, err := strconv.ParseUint(string(resByte), 10, 64)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func IncrValue(key string, value interface{}) error {
	if prefix != "" {
		key = prefix + key
	}
	if ca == nil {
		panic("cache 未初始化")
	}
	return ca.IncrValue(key, value)
}

func DecrValue(key string, value interface{}) error {
	if prefix != "" {
		key = prefix + key
	}
	if ca == nil {
		panic("cache 未初始化")
	}
	return ca.DecrValue(key, value)
}

// IsExist .
func IsExist(key string) bool {
	if prefix != "" {
		key = prefix + key
	}
	if ca == nil {
		panic("cache 未初始化")
	}
	return ca.IsExist(key)
}

// Delete .
func Delete(key string) error {
	if prefix != "" {
		key = prefix + key
	}
	if ca == nil {
		panic("cache 未初始化")
	}
	return ca.Delete(key)
}

func lock(key string) {

}
