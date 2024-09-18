package config

import (
	"encoding/json"
	"io/ioutil"
)

// DB .
type Config struct {
	Mysql    mysqlConfig `json:"mysql"`
	Redis    redisConfig `json:"redis"`
	Extra    extra       `json:"extra"`
	LogLevel string      `json:"log_level"`
}

type mysqlConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	Port     string `json:"port"`
	Prefix   string `json:"prefix"`
	Charset  string `json:"charset"`
}

type redisConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     string `json:"port"`
	Db       int    `json:"db"`
	Prefix   string `json:"prefix"`
}

type extra struct {
	DingTalkURL string `json:"ding_talk_url"`
	LoginExTime int64  `json:"login_ex_time"`
}

var internalConfig *Config

// ParseConfig .
func ParseConfig(conf *Config) {
	internalConfig = conf
}

func ParseConfigByFile(configName string) error {
	dat, err := ioutil.ReadFile(configName)
	if err != nil {
		return err
	}

	err = json.Unmarshal(dat, &internalConfig)
	if err != nil {
		return err
	}
	return nil
}

func GetConfig() *Config {
	return internalConfig
}
