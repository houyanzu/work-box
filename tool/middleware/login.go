package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/houyanzu/work-box/config"
	"github.com/houyanzu/work-box/lib/crypto"
	"github.com/houyanzu/work-box/tool/cache"
	"net/http"
	"strings"
	"time"
)

func Login() gin.HandlerFunc {
	return loginHandler
}

func AdminLogin() gin.HandlerFunc {
	return loginHandler
}

func loginHandler(c *gin.Context) {
	token := c.GetHeader("token")
	account := c.GetHeader("wallet")
	account = strings.ToLower(account)
	lang := c.GetHeader("Language")
	if token == "" {
		if lang == "zh" {
			c.JSON(http.StatusOK, gin.H{
				"code": 3,
				"msg":  "亲，登陆过期了，需要重新登录哟",
				"data": gin.H{},
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": 3,
				"msg":  "Session expired, Please Login",
				"data": gin.H{},
			})
		}
		c.Abort()
		return
	}

	userId, _ := cache.GetInt64(token)
	if userId <= 0 {
		if lang == "zh" {
			c.JSON(http.StatusOK, gin.H{
				"code": 3,
				"msg":  "亲，登陆过期了，需要重新登录哟",
				"data": gin.H{},
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": 3,
				"msg":  "Session expired, Please Login",
				"data": gin.H{},
			})
		}
		c.Abort()
		return
	}

	tokenAccount, _ := cache.GetString(token + "_address")
	if account != tokenAccount {
		if lang == "zh" {
			c.JSON(http.StatusOK, gin.H{
				"code": 3,
				"msg":  "亲，登陆过期了，需要重新登录哟",
				"data": gin.H{},
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": 3,
				"msg":  "Session expired, Please Login",
				"data": gin.H{},
			})
		}
		c.Abort()
		return
	}
	c.Set("userId", userId)
	c.Next()
	return
}

func adminLoginHandler(c *gin.Context) {
	token := c.GetHeader("token")
	tokenKey := "admin_" + token
	lang := c.GetHeader("Language")
	if token == "" {
		if lang == "zh" {
			c.JSON(http.StatusOK, gin.H{
				"code": 3,
				"msg":  "亲，登陆过期了，需要重新登录哟",
				"data": gin.H{},
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": 3,
				"msg":  "Session expired, Please Login",
				"data": gin.H{},
			})
		}
		c.Abort()
		return
	}

	userId, _ := cache.GetInt64(tokenKey)
	if userId <= 0 {
		if lang == "zh" {
			c.JSON(http.StatusOK, gin.H{
				"code": 3,
				"msg":  "亲，登陆过期了，需要重新登录哟",
				"data": gin.H{},
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": 3,
				"msg":  "Session expired, Please Login",
				"data": gin.H{},
			})
		}
		c.Abort()
		return
	}

	c.Set("userId", userId)
	c.Next()
	return
}

func GetLoginToken(userID uint, address string, alone bool) (token string, err error) {
	switchKey := fmt.Sprintf("%dlogin", userID)
	oldToken := cache.Get(switchKey)
	oldTokenStr, ok := oldToken.(string)
	if !ok {
		err = errors.New("wrong")
		return
	}
	if oldToken != "" && alone {
		cache.Delete(oldTokenStr)
	}

	had := true
	for i := 0; i < 5; i++ {
		token = crypto.Sha1Str(time.Now().Format(time.RFC3339) + fmt.Sprintf("%d", userID))
		t, _ := cache.GetString(token)
		if t == "" {
			had = false
			break
		}
	}
	if had {
		err = errors.New("had")
		return
	}

	conf := config.GetConfig()
	if alone {
		cache.Set(switchKey, token, conf.Extra.LoginExTime)
	}

	cache.Set(token, userID, conf.Extra.LoginExTime)
	cache.Set(token+"_address", strings.ToLower(address), conf.Extra.LoginExTime)
	return
}

func GetUserId(c *gin.Context) uint {
	userIdInterface, _ := c.Get("userId")
	userIdInt, _ := userIdInterface.(int64)
	userId := uint(userIdInt)
	return userId
}
