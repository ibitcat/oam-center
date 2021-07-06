package router

import (
	"strconv"
	"strings"
	"time"

	"oam-center/libs"
	"oam-center/models"

	"github.com/gin-gonic/gin"
)

func PersonalIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "资料修改",
	}

	userId := getUserId(c)
	user := models.UserGetById(userId)
	if user != nil {
		row := make(gin.H)
		row["id"] = user.Id
		row["login_name"] = user.LoginName
		row["real_name"] = user.RealName
		row["phone"] = user.Phone
		row["email"] = user.Email
		data["admin"] = row
	}

	c.HTML(200, "personal", data)
}

func PersonalAjaxSave(c *gin.Context) {
	userId, _ := strconv.Atoi(c.PostForm("id"))
	user := models.UserGetById(userId)
	if user != nil {
		c.Bind(user)

		resetPwd := c.PostForm("reset_pwd")
		if resetPwd == "1" {
			pwdOld := strings.TrimSpace(c.PostForm("password_old"))
			pwdOldMd5 := libs.Md5([]byte(pwdOld + user.Salt))
			if user.Password != pwdOldMd5 {
				c.JSON(200, gin.H{"status": -1, "message": "旧密码错误"})
				return
			}

			pwdNew1 := strings.TrimSpace(c.PostForm("password_new1"))
			pwdNew2 := strings.TrimSpace(c.PostForm("password_new2"))
			if len(pwdNew1) < 6 {
				c.JSON(200, gin.H{"status": -1, "message": "密码长度需要六位以上"})
				return
			}

			if pwdNew1 != pwdNew2 {
				c.JSON(200, gin.H{"status": -1, "message": "两次密码不一致"})
				return
			}

			pwd, salt := libs.Password(4, pwdNew1)
			user.Password = pwd
			user.Salt = salt
			if err := user.Update(); err != nil {
				c.JSON(200, gin.H{"status": -1, "message": err.Error()})
			} else {
				c.JSON(200, gin.H{"status": 0})
			}
		} else {
			user.UpdateId = userId
			user.UpdateTime = time.Now().Unix()
			user.Update()
			c.JSON(200, gin.H{"status": 0})
		}
	}
}
