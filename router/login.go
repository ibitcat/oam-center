package router

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"oam-center/libs"
	"oam-center/models"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

const SiteName = "xx运维后台 v1.0"

func LoginIndex(c *gin.Context) {
	data := gin.H{"siteName": SiteName}
	c.HTML(200, "login", data)
}

func Login(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := strings.TrimSpace(c.PostForm("password"))
	if len(username) == 0 || len(password) == 0 {
		c.AbortWithStatusJSON(200, gin.H{"status": -1, "message": "请输入账号密码"})
		return
	} else {
		user := models.UserGetByName(username)
		if user == nil {
			c.AbortWithStatusJSON(200, gin.H{"status": -1, "message": "用户不存在"})
			return
		} else if user.Password != libs.Md5([]byte(password+user.Salt)) {
			c.AbortWithStatusJSON(200, gin.H{"status": -1, "message": "帐号或密码错误"})
			return
		} else if user.Status == -1 {
			c.AbortWithStatusJSON(200, gin.H{"status": -1, "message": "该帐号已禁用"})
			return
		} else {
			ip := getClientIp(c)
			user.LastIp = ip
			user.LastLogin = time.Now().Unix()
			user.Update("last_ip", "last_login")

			// session
			session := sessions.Default(c)
			authkey := libs.Md5([]byte(ip + "|" + user.Password + user.Salt))
			cookie := strconv.Itoa(user.Id) + "|" + user.LoginName + "|" + authkey
			session.Set("auth", cookie)
			session.Save()

			c.JSON(200, gin.H{"status": 0})
		}
		fmt.Println(user)
	}
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("auth")

	c.Redirect(301, "/")
}
