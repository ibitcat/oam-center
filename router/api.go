package router

import (
	"fmt"
	"strings"
	"time"

	"oam-center/conf"
	"oam-center/models"

	"github.com/gin-gonic/gin"
)

type CtlConf struct {
	Ip      string
	Detail  string
	Domain  string
	Version string
	Type    uint
	VpsTime int64
}

func VpsOnlineApi(c *gin.Context) {
	ctl := CtlConf{}
	err := c.BindJSON(&ctl)
	if err == nil {
		if len(ctl.Ip) > 0 {
			vps := models.VpsGetByIp(ctl.Ip)
			if vps == nil {
				vps = &models.Vps{}
				vps.Ip = ctl.Ip
				vps.Type = ctl.Type
				vps.Detail = ctl.Detail
				vps.Domain = ctl.Domain
				vps.CreateTime = time.Now().Unix()
				vps.Status = 1
				vps.CtlVersion = ctl.Version
				vps.VpsTime = ctl.VpsTime
				vps.Insert()
			} else {
				if ctl.Detail != vps.Detail {
					vps.Detail = ctl.Detail
				}
				if vps.Type != ctl.Type {
					vps.Type = ctl.Type
				}
				if vps.Domain != ctl.Domain {
					vps.Domain = ctl.Domain
				}
				if vps.VpsTime != ctl.VpsTime {
					vps.VpsTime = ctl.VpsTime
				}

				vps.CtlVersion = ctl.Version
				vps.Status = 1
				vps.Update()
			}
		}
	}
	c.JSON(200, gin.H{"status": 0, "message": "连接到oam-center成功"})
}

func ProcsUpdateApi(c *gin.Context) {
	procs := make(map[string]string)
	err := c.BindJSON(&procs)
	if err == nil {
		push := make([]string, 0, len(procs))
		for gameFlag, v := range procs {
			game := models.GameGetByFlag(gameFlag)
			if game != nil {
				if game.Procs != v {
					game.Procs = v
					game.Update("procs")

					push = append(push, fmt.Sprintf("游服[%s]进程有变化，进程列表：[%s]", gameFlag, v))
				}
			}
		}

		// 钉钉推送
		if len(push) > 0 {
			data := gin.H{
				"msgtype": "text",
				"text":    gin.H{"content": strings.Join(push, "\n")},
			}
			models.HttpPostJson(conf.YamlConf.App.DDpush, data)
		}
	}

	c.JSON(200, gin.H{"status": 0, "message": "更新进程信息"})
}

func UpdateNoticeApi(c *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0})
		}
	}()

	n := &models.CDNNotice{}
	err = c.BindJSON(n)
	if err != nil {
		return
	}

	if len(n.Title) == 0 || len(n.Content) == 0 {
		err = fmt.Errorf("请输入公告标题和公告内容")
		return
	}

	s := strings.Split(n.Version, "_")
	if len(s) != 3 {
		err = fmt.Errorf("CDN版本格式错误")
		return
	}

	cdn := models.CDNGetByVersion(n.Version)
	if n.Debug == 0 && (cdn == nil || cdn.Status == 0) {
		err = fmt.Errorf("CDN版本不存在或未安装")
		return
	}

	agent := models.AgentGetByFlag(s[1])
	if agent == nil {
		err = fmt.Errorf("CDN所属平台不存在")
	} else {
		err = agent.UpdateNotice(n, s[2])
	}
}
