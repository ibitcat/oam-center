package router

import (
	"fmt"
	"html/template"
	"strconv"
	"time"

	"oam-center/libs"
	"oam-center/models"

	"github.com/gin-gonic/gin"
)

var NginxStatusTxt = []string{"<font color='red'>未安装</font>", "已安装"}

func NginxIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "nginx管理",
	}
	c.HTML(200, "nginxIndex", data)
}

func NginxAddIndex(c *gin.Context) {
	userName := getUserName(c)

	// 所有小程序平台
	agents, _ := models.AgentGetList("miniapp", 1)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "nginx注册",
		"agentGroup":    agents,
	}
	c.HTML(200, "nginxAdd", data)
}

func NginxTable(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))

	nginxs, count := models.NginxGetPageList(page, limit)
	list := make([]gin.H, 0, len(nginxs))
	for _, v := range nginxs {
		agent := models.AgentGetByAid(v.Aid)
		row := make(gin.H)
		row["id"] = v.Id
		row["aid"] = v.Aid
		row["agent"] = agent.Flag
		row["sid"] = v.Sid
		row["domain"] = v.Domain
		row["ws"] = v.Ws
		row["single"] = v.Single
		row["vps"] = v.Vpsid
		row["create_time"] = libs.FormatTime(v.CreateTime)
		row["status_txt"] = template.HTML(NginxStatusTxt[v.Status])
		list = append(list, row)
	}
	c.JSON(200, gin.H{"code": 0, "count": count, "data": list})
}

func checkNginx(nginx *models.Nginx) error {
	aid, sid := nginx.Aid, nginx.Sid
	if aid == 0 || sid == 0 {
		return fmt.Errorf("请选择平台和起始sid")
	}

	nginxs, _ := models.NginxGetList("aid", aid, "sid", sid)
	if len(nginxs) != 0 {
		return fmt.Errorf("该nginx转发服务器已存在")
	}

	// 检查vps
	vps := models.VpsGetById(nginx.Vpsid)
	if vps == nil {
		return fmt.Errorf("vps不存在")
	}

	// vps是否是Nginx
	if !vps.CheckType(models.VpsNginx) {
		return fmt.Errorf("该vps非Nginx类型")
	}
	return nil
}

func NginxAjaxSave(c *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0})
		}
	}()

	id, _ := strconv.Atoi(c.PostForm("id"))
	if id == 0 {
		nginx := new(models.Nginx)
		c.Bind(nginx)

		err = checkNginx(nginx)
		if err != nil {
			return
		}

		nginx.Status = 0
		nginx.Ws = "wss-" + nginx.Domain
		nginx.Single = "single-" + nginx.Domain
		nginx.CreateTime = time.Now().Unix()
		nginx.Insert()
	} else {
		nginx := models.NginxGetById(id)
		if nginx == nil {
			err = fmt.Errorf("Nginx转发不存在")
			return
		} else {
			c.Bind(nginx)
		}

		err = checkNginx(nginx)
		if err != nil {
			return
		}
		nginx.Ws = "wss-" + nginx.Domain
		nginx.Single = "single-" + nginx.Domain
		nginx.Update()
	}
}

func NginxUpdateDomain(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	domain := c.DefaultPostForm("domain", "")
	if id == 0 {
		c.JSON(200, gin.H{"status": -1, "message": "Nginx转发不存在"})
	} else {
		if len(domain) == 0 {
			c.JSON(200, gin.H{"status": -1, "message": "转发域名不能为空"})
		} else {
			nginx := models.NginxGetById(id)
			if nginx != nil {
				nginx.Domain = domain
				nginx.Ws = "wss-" + domain
				nginx.Single = "single-" + domain
				nginx.Update()
			}
			c.JSON(200, gin.H{"status": 0})
		}
	}
}

func NginxInstall(c *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0})
		}
	}()

	id, _ := strconv.Atoi(c.PostForm("id"))
	nginx := models.NginxGetById(id)
	if nginx != nil {
		if nginx.Status > 0 {
			err = fmt.Errorf("Nginx转发服务器已安装")
			return
		}

		err = nginx.Install()
	} else {
		err = fmt.Errorf("Nginx转发不存在")
	}
}
