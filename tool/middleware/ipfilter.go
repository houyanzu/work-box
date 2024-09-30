package middleware

import (
	"exchange_pro/lib/tool"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/houyanzu/work-box/config"
	"github.com/houyanzu/work-box/tool/cache"
	"sync"
	"tang/lib/output"
)

var mu = sync.Mutex{}

func IPFilter() gin.HandlerFunc {
	return ipFilterHandler
}

func ipFilterHandler(c *gin.Context) {
	conf := config.GetConfig()
	if !conf.IPFilter.Enable {
		c.Next()
		return
	}
	clientIP := c.ClientIP()
	if tool.InSlice(clientIP, conf.IPFilter.WhiteList) {
		c.Next()
		return
	}
	if tool.InSlice(clientIP, conf.IPFilter.BlackList) {
		output.NewOutput(c, 0, nil).Out()
		c.Abort()
		return
	}
	for _, rule := range conf.IPFilter.Rules {
		if !rule.Enable {
			continue
		}
		if rule.Duration == 0 {
			continue
		}
		key := fmt.Sprintf("ip_filter_%s_%s_times", rule.Module, clientIP)
		if !cache.IsExist(key) {
			func() {
				mu.Lock()
				defer mu.Unlock()
				if !cache.IsExist(key) {
					_ = cache.Set(key, 1, rule.Duration)
				}
			}()
		}
		times, _ := cache.GetInt64(key)
		if times > rule.Threshold {
			output.NewOutput(c, 0, nil).Out()
			c.Abort()
			return
		}
		_ = cache.IncrValue(key, 1)
	}

	perpetualKey := fmt.Sprintf("ip_filter_%s_%s", "perpetual", clientIP)
	if cache.IsExist(perpetualKey) {
		output.NewOutput(c, 0, nil).Out()
		c.Abort()
		return
	}

	perpetualTimesKey := fmt.Sprintf("ip_filter_%s_%s_times", "perpetual", clientIP)
	if !cache.IsExist(perpetualTimesKey) {
		func() {
			mu.Lock()
			defer mu.Unlock()
			if !cache.IsExist(perpetualTimesKey) {
				_ = cache.Set(perpetualTimesKey, 1, 3600)
			}
		}()
	}
	perpetualTimes, _ := cache.GetInt64(perpetualTimesKey)
	if perpetualTimes > conf.IPFilter.PerpetualTimesAHour {
		_ = cache.Set(perpetualKey, 1, 0)
		output.NewOutput(c, 0, nil).Out()
		c.Abort()
		return
	}
	_ = cache.IncrValue(perpetualTimesKey, 1)

	if !conf.IPFilter.TokenFilter {
		c.Next()
		return
	}

	userToken := c.GetHeader("token")
	if userToken == "" || userToken == "null" || userToken == "undefined" {
		c.Next()
		return
	}

	for _, rule := range conf.IPFilter.Rules {
		if !rule.Enable {
			continue
		}
		if rule.Duration == 0 {
			continue
		}
		key := fmt.Sprintf("ip_filter_%s_%s_times", rule.Module, userToken)
		if !cache.IsExist(key) {
			func() {
				mu.Lock()
				defer mu.Unlock()
				if !cache.IsExist(key) {
					_ = cache.Set(key, 1, rule.Duration)
				}
			}()
		}
		times, _ := cache.GetInt64(key)
		if times > rule.Threshold {
			output.NewOutput(c, 0, nil).Out()
			c.Abort()
			return
		}
		_ = cache.IncrValue(key, 1)
	}

	c.Next()
}
