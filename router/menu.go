package router

import (
	"strconv"
	"time"

	"oam-center/models"

	"github.com/gin-gonic/gin"
)

func MenuIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"zTree":         true,
		"pageTitle":     "权限因子",
	}
	c.HTML(200, "menuIndex", data)
}

//获取全部节点
func GetNodes(c *gin.Context) {
	menus, count := models.MenuGetList("status", 1)
	list := make([]gin.H, 0, len(menus))
	for _, v := range menus {
		row := make(gin.H)
		row["id"] = v.Id
		row["pId"] = v.Pid
		row["name"] = v.AuthName
		row["open"] = true
		list = append(list, row)
	}

	c.JSON(200, gin.H{"code": 0, "count": count, "data": list})
}

//获取一个节点
func GetNode(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	menu := models.MenuGetById(id)
	if menu != nil {
		row := make(gin.H)
		row["id"] = menu.Id
		row["pid"] = menu.Pid
		row["auth_name"] = menu.AuthName
		row["auth_url"] = menu.AuthUrl
		row["sort"] = menu.Sort
		row["is_show"] = menu.IsShow
		row["icon"] = menu.Icon

		c.JSON(200, gin.H{"code": 0, "count": 0, "data": row})
	}
}

func MenuAjaxSave(c *gin.Context) {
	userId := getUserId(c)
	menuId, _ := strconv.Atoi(c.PostForm("id"))
	if menuId == 0 {
		menu := new(models.Menu)
		c.Bind(menu)

		menu.Status = 1
		menu.CreateTime = time.Now().Unix()
		menu.CreateId = userId
		menu.UpdateId = userId
		if err := menu.Insert(); err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0})
		}
	} else {
		menu := models.MenuGetById(menuId)
		if menu == nil {
			c.JSON(200, gin.H{"status": -1, "message": "菜单不存在"})
		} else {
			c.Bind(menu)
		}

		menu.UpdateId = userId
		menu.UpdateTime = time.Now().Unix()
		menu.Update()
		c.JSON(200, gin.H{"status": 0})
	}
}

func MenuAjaxDel(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	menu := models.MenuGetById(id)
	if menu != nil {
		menu.Status = 0
		menu.UpdateId = getUserId(c)
		menu.UpdateTime = time.Now().Unix()
		menu.Update()
		c.JSON(200, gin.H{"status": 0})
	}
}
