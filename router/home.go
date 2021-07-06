package router

import (
	//"log"
	"net/http"

	"oam-center/libs"
	"oam-center/models"

	"github.com/gin-gonic/gin"
)

func HomeIndex(c *gin.Context) {
	userId := getUserId(c)
	userName := getUserName(c)

	user := models.UserGetById(userId)
	menus := models.MenuGetAuth(user.RoleId)
	list := make([]gin.H, 0, len(menus))
	list2 := make([]gin.H, 0, len(menus))
	for _, v := range menus {
		row := make(gin.H)
		if v.Pid == 1 && v.IsShow == 1 {
			// 一级菜单
			row["Id"] = int(v.Id)
			row["Sort"] = v.Sort
			row["AuthName"] = v.AuthName
			row["AuthUrl"] = v.AuthUrl
			row["Icon"] = v.Icon
			row["Pid"] = int(v.Pid)
			list = append(list, row)
		}

		if v.Pid != 1 && v.IsShow == 1 {
			//二级菜单
			row["Id"] = int(v.Id)
			row["Sort"] = v.Sort
			row["AuthName"] = v.AuthName
			row["AuthUrl"] = v.AuthUrl
			row["Icon"] = v.Icon
			row["Pid"] = int(v.Pid)
			list2 = append(list2, row)
		}
	}

	data := gin.H{
		"title":         SiteName,
		"siteName":      SiteName,
		"loginUserName": userName,
		"SideMenu1":     list,
		"SideMenu2":     list2,
	}
	c.HTML(200, "home", data)
}

func HomeStartIndex(c *gin.Context) {
	userName := getUserName(c)

	gsCount := models.GameGetCount()
	agentCount := models.AgentGetCount()
	cdnCount := models.CDNGetCount()

	// 最近新开20个的服务器
	lastSvrs := models.GameGetNearList()
	nearGss := make([]gin.H, 0, len(lastSvrs))
	for _, v := range lastSvrs {
		row := make(gin.H)
		row["id"] = v.Id
		row["name"] = v.Name
		row["flag"] = v.GetFlag()
		row["version"] = v.Version
		row["createTime"] = libs.FormatTime(v.CreateTime)
		nearGss = append(nearGss, row)
	}

	// 所有的版本
	vers := GetAllVersions()
	nearVers := make([]gin.H, 0, 20)
	for i, v := range vers {
		if i >= cap(nearVers) {
			break
		}
		nearVers = append(nearVers, gin.H{
			"name": v.Name(),
			"time": v.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	//系统运行信息
	info := libs.SystemInfo(models.StartTime)

	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "系统概况",
		"gameAmount":    gsCount,
		"agentAmount":   agentCount,
		"cdnAmount":     cdnCount,
		"versionAmount": len(vers),
		"nearGsList":    nearGss,
		"nearVers":      nearVers,
		"sysInfo":       info,
	}

	// 系统概况
	c.HTML(http.StatusOK, "start", data)
}
