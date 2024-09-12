package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/houyanzu/work-box/config"
	"github.com/houyanzu/work-box/lib/crypto"
	"github.com/houyanzu/work-box/lib/output"
	"github.com/houyanzu/work-box/tool/cache"
	"strings"
	"time"
)

func Login() gin.HandlerFunc {
	return loginHandler
}

func AdminLogin() gin.HandlerFunc {
	return adminLoginHandler
}

func loginHandler(c *gin.Context) {
	token := c.GetHeader("token")
	account := c.GetHeader("wallet")
	account = strings.ToLower(account)
	if token == "" {
		output.NewOutput(c, 3, nil).Out()
		c.Abort()
		return
	}

	userId, _ := cache.GetInt64(token)
	if userId <= 0 {
		output.NewOutput(c, 3, nil).Out()
		c.Abort()
		return
	}

	tokenAccount, _ := cache.GetString(token + "_address")
	if account != tokenAccount {
		output.NewOutput(c, 3, nil).Out()
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
	if token == "" {
		output.NewOutput(c, 3, nil).Out()
		c.Abort()
		return
	}

	userId, _ := cache.GetInt64(tokenKey)
	if userId <= 0 {
		output.NewOutput(c, 3, nil).Out()
		c.Abort()
		return
	}

	c.Set("userId", userId)
	c.Next()
	return
}

func GetLoginToken(userID uint, address string, alone bool) (token string, err error) {
	switchKey := fmt.Sprintf("%dlogin", userID)
	oldToken, _ := cache.GetString(switchKey)
	//if err != nil {
	//	return
	//}
	if oldToken != "" && alone {
		cache.Delete(oldToken)
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

func GetAdminLoginToken(userId uint) string {
	userIdStr := fmt.Sprintf("%d", userId)
	oldTokenExist := cache.IsExist(userIdStr + "admin_login")
	if oldTokenExist {
		oldTokenByte, _ := cache.Get(userIdStr + "admin_login").([]byte)
		_ = cache.Delete("admin_" + string(oldTokenByte))
		_ = cache.Delete(userIdStr + "admin_login")
	}

	token := ""
	had := false
	for i := 0; i < 5; i++ {
		token = crypto.Sha1Str(fmt.Sprintf("%d", time.Now().UnixNano()) + userIdStr)
		had = cache.IsExist(token)
		if !had {
			break
		}
	}

	if had {
		return ""
	}

	_ = cache.Set(userIdStr+"admin_login", token, 3600)
	_ = cache.Set("admin_"+token, userId, 3600)
	return token
}

func GetUserId(c *gin.Context) uint {
	userIdInterface, _ := c.Get("userId")
	userIdInt, _ := userIdInterface.(int64)
	userId := uint(userIdInt)
	return userId
}
