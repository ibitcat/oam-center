package router

import (
	"log"
	"strconv"
	"strings"
	"time"

	"oam-center/libs"
	"oam-center/models"

	"github.com/gin-gonic/gin"
)

var UserStatusText = map[int]string{
	0: "<font color='red'>禁用</font>",
	1: "正常",
}

func UserIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "用户管理",
	}
	c.HTML(200, "userIndex", data)
}

func UserAddIndex(c *gin.Context) {
	// 角色列表
	roles, _ := models.RoleGetList("status", 1)
	list := make([]gin.H, 0, len(roles))
	for _, v := range roles {
		row := make(gin.H)
		row["id"] = v.Id
		row["role_name"] = v.RoleName
		list = append(list, row)
	}

	data := gin.H{
		"role":      list,
		"pageTitle": "新增管理员",
	}
	c.HTML(200, "userAdd", data)
}

func UserEditIndex(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Query("id"))
	user := models.UserGetById(userId)
	if user != nil {
		row := make(gin.H)
		row["id"] = user.Id
		row["login_name"] = user.LoginName
		row["real_name"] = user.RealName
		row["phone"] = user.Phone
		row["email"] = user.Email
		row["role_id"] = user.RoleId

		roles, _ := models.RoleGetList("status", 1)
		list := make([]gin.H, 0, len(roles))
		for _, v := range roles {
			row := make(gin.H)
			row["id"] = v.Id
			row["role_name"] = v.RoleName
			row["checked"] = 0
			if v.Id == user.RoleId {
				row["checked"] = 1
			}

			list = append(list, row)
		}
		data := gin.H{
			"role":      list,
			"admin":     row,
			"pageTitle": "编辑用户",
		}
		c.HTML(200, "userEdit", data)
	}
}

func UserAjaxSave(c *gin.Context) {
	curUserId := getUserId(c)

	userId, _ := strconv.Atoi(c.PostForm("id"))
	if userId == 0 {
		user := new(models.User)
		c.Bind(user)

		user.UpdateTime = time.Now().Unix()
		user.UpdateId = curUserId
		user.Status = 1

		oldUser := models.UserGetByName(user.LoginName)
		if oldUser != nil {
			c.JSON(200, gin.H{"status": -1, "message": "登录名已经存在"})
			return
		}
		pwd, salt := libs.Password(4, "")
		user.Password = pwd
		user.Salt = salt
		user.CreateTime = time.Now().Unix()
		user.CreateId = curUserId

		if err := user.Insert(); err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0})
		}
	} else {
		user := models.UserGetById(userId)
		if user == nil {
			c.JSON(200, gin.H{"status": -1, "message": "用户不存在"})
			return
		} else {
			c.Bind(user)
			log.Println(user)
		}

		//普通管理员不可修改超级管理员资料
		if curUserId != 1 && user.Id == 1 {
			c.JSON(200, gin.H{"status": -1, "message": "不可修改超级管理员资料"})
			return
		}

		user.UpdateTime = time.Now().Unix()
		if err := user.Update(); err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0})
		}
	}
}

func UserAjaxDel(c *gin.Context) {
	userId, _ := strconv.Atoi(c.PostForm("id"))
	status := c.PostForm("status")
	user := models.UserGetById(userId)
	if user != nil {
		if user.Id == 1 {
			c.JSON(200, gin.H{"status": -1, "message": "超级管理员不允许操作"})
			return
		}

		user.UpdateTime = time.Now().Unix()
		user.Status = 0
		if status == "enable" {
			user.Status = 1
		}

		if err := user.Update("update_time", "status"); err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0})
		}
	}
}

func UserTable(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))
	realName := strings.TrimSpace(c.DefaultQuery("realName", ""))

	var count int
	var users []models.User
	if len(realName) > 0 {
		users, count = models.UserGetPageList(page, limit, "real_name", realName)
	} else {
		users, count = models.UserGetPageList(page, limit)
	}

	list := make([]gin.H, 0, len(users))
	for _, v := range users {
		row := make(gin.H)
		row["id"] = v.Id
		row["login_name"] = v.LoginName
		row["real_name"] = v.RealName
		row["phone"] = v.Phone
		row["email"] = v.Email
		row["role_id"] = v.RoleId
		row["create_time"] = libs.FormatTime(v.CreateTime)
		row["update_time"] = libs.FormatTime(v.UpdateTime)
		row["status"] = v.Status
		row["status_text"] = UserStatusText[v.Status]
		list = append(list, row)
	}
	c.JSON(200, gin.H{"code": 0, "count": count, "data": list})
}
