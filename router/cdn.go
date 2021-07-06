package router

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"oam-center/libs"
	"oam-center/models"

	"github.com/gin-gonic/gin"
)

var CDNStatusTxt = []string{
	"<font color='red'><i class='fa fa-minus-square'></i>未安装</font>",
	"<font color='green'><i class='fa fa-check-square'></i>已安装</font>",
	"<font color='orange'><i class='fa fa-check-square'></i>已删除</font>",
	"<font color='gray'><i class='fa fa-check-square'></i>已停止</font>",
}

// 数据库版本是否存在于硬盘（删除）
func isDBInDisk(cdnVer string, vers []string) bool {
	for _, ver := range vers {
		if cdnVer == ver {
			return true
		}
	}
	return false
}

// 硬盘上的版本是否存在于数据库（增加）
func isDiskInDB(ver string, cdns []models.CDN) bool {
	for _, cdn := range cdns {
		if cdn.Version == ver {
			return true
		}
	}
	return false
}

func initAllCdnVersion() {
	agents, _ := models.AgentGetList("status", 1)
	for _, agent := range agents {
		// 该平台所有cdn版本源文件
		vers := GetVersions(agent.Flag, 0)

		// 该平台下的所有cdn版本
		cdns, _ := models.CDNGetList("aid", agent.Aid)
		for _, cdn := range cdns {
			if !isDBInDisk(cdn.Version, vers) {
				cdn.Status = 2 // 版本源文件已删除
				cdn.Update("status")
			}
		}

		for _, ver := range vers {
			if !isDiskInDB(ver, cdns) {
				cdn := &models.CDN{}
				cdn.Aid = agent.Aid
				cdn.Version = ver
				cdn.Status = 0
				cdn.CreateTime = time.Now().Unix()
				cdn.Insert()
			}
		}
	}
}

func CDNIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "CND管理",
	}

	c.HTML(200, "cdnIndex", data)
}

func CDNNoticeIndex(c *gin.Context) {
	userName := getUserName(c)

	id, _ := strconv.Atoi(c.Query("id"))
	cdn := models.CDNGetById(id)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "公告编辑",
		"cdn":           cdn,
	}
	c.HTML(200, "cdnNotice", data)
}

func CDNFreshVersions(c *gin.Context) {
	initAllCdnVersion()
	c.JSON(200, gin.H{"status": 0, "message": "刷新所有CDN版本成功"})
}

func CDNCleanUp(c *gin.Context) {
	models.CDNDelete()
	c.JSON(200, gin.H{"status": 0, "message": "清理已停止的CDN成功"})
}

func CDNTable(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))

	var amount int
	var cdns []models.CDN
	// cdnVersion = cn_andou_20190101
	cdnVersion := strings.TrimSpace(c.DefaultQuery("cdn_version", ""))
	if len(cdnVersion) > 0 {
		cdns, amount = models.CDNGetPageList(page, limit, "version", cdnVersion)
	} else {
		cdns, amount = models.CDNGetPageList(page, limit)
	}

	list := make([]gin.H, 0, len(cdns))

	for _, v := range cdns {
		agent := models.AgentGetByAid(v.Aid)
		if agent != nil {
			row := make(gin.H)
			row["id"] = v.Id
			row["aid"] = v.Aid
			row["agent_flag"] = agent.GetFlag()
			row["vpsid"] = agent.Vpsid
			row["version"] = v.Version
			row["source"] = agent.Source
			row["domain"] = agent.Domain
			row["create_time"] = libs.FormatTime(v.CreateTime)
			row["status_txt"] = template.HTML(CDNStatusTxt[v.Status])
			list = append(list, row)
		}
	}
	c.JSON(200, gin.H{"code": 0, "count": amount, "data": list})
}

func CDNInstall(c *gin.Context) {
	tags := c.PostForm("tags")
	isTest := c.DefaultPostForm("test", "")
	if len(tags) > 0 {
		var msg string
		cdn := models.CDNGetByTag(tags)
		if cdn == nil {
			msg = "cdn不存在"
		} else {
			agent := models.AgentGetByAid(cdn.Aid)
			if agent == nil {
				msg = "cdn平台不存在"
			} else {
				vpsDomain := models.VpsGetDomain(agent.Vpsid)
				data := gin.H{
					"tag":    cdn.Version,
					"source": agent.Source,
					"domain": agent.Domain,
					"isTest": 0,
				}
				if len(isTest) > 0 {
					data["isTest"] = 1
				}

				apiUrl := fmt.Sprintf("%s/cdn/install", vpsDomain)
				rsp, err := models.HttpPostJson(apiUrl, data)
				if err == nil {
					if rsp.Status == 0 {
						msg = "cdn安装成功"
						cdn.Status = 1
					} else {
						msg = "cdn安装失败，详情请查看日志!"
					}
					if len(isTest) == 0 {
						cdn.InstallLog = rsp.Message
						cdn.Update("status", "install_log")
					}
				} else {
					msg = "cdn安装失败" + err.Error()
				}
			}
		}
		c.JSON(200, gin.H{"status": 0, "message": msg})
	} else {
		c.JSON(200, gin.H{"status": -1, "message": "请先选择要操作的cdn！"})
	}
}

func CDNInstallTest(c *gin.Context) {
	tag := c.PostForm("tag")
	if len(tag) > 0 {
		var msg string
		cdn := models.CDNGetByTag(tag)
		if cdn == nil {
			msg = "cdn不存在"
		} else {
			agent := models.AgentGetByAid(cdn.Aid)
			if agent == nil {
				msg = "cdn平台不存在"
			} else {
				vpsDomain := models.VpsGetDomain(agent.Vpsid)
				data := gin.H{
					"tag":    cdn.Version,
					"source": agent.Source,
					"domain": agent.Domain,
				}

				apiUrl := fmt.Sprintf("%s/cdn/install", vpsDomain)
				rsp, err := models.HttpPostJson(apiUrl, data)
				if err == nil {
					if rsp.Status == 0 {
						msg = "cdn安装成功"
						cdn.Status = 1
					} else {
						msg = "cdn安装失败，详情请查看日志!"
					}
					cdn.InstallLog = rsp.Message
					cdn.Update("status", "install_log")
				} else {
					msg = "cdn安装失败" + err.Error()
				}
			}
		}
		c.JSON(200, gin.H{"status": 0, "message": msg})
	} else {
		c.JSON(200, gin.H{"status": -1, "message": "请先选择要操作的cdn！"})
	}
}

func CDNAjaxLog(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	cdn := models.CDNGetById(id)
	if cdn != nil {
		c.JSON(200, gin.H{"status": 0, "message": cdn.InstallLog})
	} else {
		c.JSON(200, gin.H{"status": -1, "message": "cdn不存在"})
	}
}

// 更新某个版本的公告到cdn（eg.: cn_andou/v20190101/notice.json）
func CDNUpdateNotice(c *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0})
		}
	}()

	id, _ := strconv.Atoi(c.PostForm("id"))
	n := &models.CDNNotice{}
	err = c.Bind(n)
	if err != nil {
		return
	}

	if len(n.Title) == 0 || len(n.Content) == 0 {
		err = fmt.Errorf("请输入公告标题和公告内容")
		return
	}

	cdn := models.CDNGetById(id)
	if n.Debug == 0 && (cdn == nil || cdn.Status == 0) {
		err = fmt.Errorf("CDN版本不存在或未安装")
		return
	}

	s := strings.Split(cdn.Version, "_")
	if len(s) != 3 {
		err = fmt.Errorf("CDN版本格式错误")
		return
	}

	agent := models.AgentGetByAid(cdn.Aid)
	if agent == nil {
		err = fmt.Errorf("CDN所属平台不存在")
	} else {
		err = agent.UpdateNotice(n, s[2])
	}
}

func CDNStopRemote(c *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0, "message": "停止解析CND并删除成功"})
		}
	}()

	id, _ := strconv.Atoi(c.PostForm("id"))
	cdn := models.CDNGetById(id)
	if cdn != nil {
		if cdn.Status != 2 {
			err = fmt.Errorf("无法停止该CDN")
			return
		}

		agent := models.AgentGetByAid(cdn.Aid)
		if agent == nil {
			err = fmt.Errorf("CDN所属平台不存在")
		} else {
			err = agent.StopCDN(cdn.Version)
			if err == nil {
				cdn.Status = 3
				cdn.Update("status")
			}
		}
	}
}
