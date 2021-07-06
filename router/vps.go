package router

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"oam-center/libs"
	"oam-center/models"

	"github.com/gin-gonic/gin"
)

var VpsStatusTxt = []string{"离线", "在线"}
var VpsTypeTxt = map[uint]string{
	models.VpsGame:   "游服",
	models.VpsCDN:    "CDN",
	models.VpsNginx:  "Nginx",
	models.VpsCenter: "中央后台",
	models.VpsCross:  "跨服",
}

func VpsIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "Vps列表",
	}
	c.HTML(200, "vpsIndex", data)
}

func VpsTable(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))

	vpss, count := models.VpsGetPageList(page, limit)
	list := make([]gin.H, 0, len(vpss))
	for _, v := range vpss {
		row := make(gin.H)
		row["id"] = v.Id
		row["ip"] = v.Ip
		row["type"] = getVpsTypeString(v.Type)
		row["domain"] = v.Domain
		row["detail"] = v.Detail
		row["status_txt"] = VpsStatusTxt[v.Status]
		row["version"] = v.CtlVersion
		row["create_time"] = libs.FormatTime(v.CreateTime)
		row["vps_time"] = libs.FormatTime(v.VpsTime)
		list = append(list, row)
	}
	c.JSON(200, gin.H{"code": 0, "count": count, "data": list})
}

func getVpsTypeString(ty uint) string {
	s := make([]string, 0, len(VpsTypeTxt))
	for k, v := range VpsTypeTxt {
		if k&ty == k {
			s = append(s, v)
		}
	}
	return strings.Join(s, ",")
}

func VpsPing(c *gin.Context) {
	id, _ := strconv.Atoi(c.DefaultQuery("id", "0"))
	vps := models.VpsGetById(id)
	if vps == nil {
		c.JSON(200, gin.H{"status": -1, "message": "云服务器不存在"})
	} else {
		err := models.PingVps(vps)
		if err == nil {
			c.JSON(200, gin.H{"status": 0, "message": "vps在线"})
		} else {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		}
	}
}

func VpsBatchPing(c *gin.Context) {
	idstr := c.DefaultPostForm("ids", "")
	if len(idstr) > 0 {
		ids := strings.Split(idstr, ",")

		var wg sync.WaitGroup
		for _, v := range ids {
			id, _ := strconv.Atoi(v)
			if id > 0 {
				vps := models.VpsGetById(id)
				if vps != nil {
					wg.Add(1)
					go func(v *models.Vps) {
						defer wg.Done()
						models.PingVps(v)
					}(vps)
				}
			}
		}
		wg.Wait()
		c.JSON(200, gin.H{"status": 0, "message": "批量ping vps完成"})
	} else {
		c.JSON(200, gin.H{"status": -1, "message": "请选择vps列表"})
	}
}

func VpsUpdateCtl(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	vps := models.VpsGetById(id)
	if vps != nil {
		apiUrl := fmt.Sprintf("%s/%s", vps.Domain, "updatectl")
		resp, err := http.Get(apiUrl)
		if err != nil {
			return
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0, "message": string(body)})
		}
	} else {
		c.JSON(200, gin.H{"status": -1, "message": "云服务器不存在"})
	}
}

func VpsChangeTime(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	timeStr := c.PostForm("timestr")
	vpsTime := libs.ParseTime(timeStr)
	if vpsTime == 0 {
		c.JSON(200, gin.H{"status": -1, "message": "时间格式错误"})
		return
	}

	vps := models.VpsGetById(id)
	if vps != nil {
		apiUrl := fmt.Sprintf("%s/%s", vps.Domain, "changetime")
		rsp, err := models.HttpPostJson(apiUrl, gin.H{"timestr": timeStr})
		if err == nil && rsp.Status == 0 {
			vps.VpsTime = vpsTime
			vps.Update("vps_time")
			c.JSON(200, gin.H{"status": 0})
		} else {
			var msg string
			if err != nil {
				msg = err.Error()
			} else {
				msg = rsp.Message
			}
			c.JSON(200, gin.H{"status": -1, "message": msg})
		}
	}
}
