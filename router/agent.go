package router

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"oam-center/libs"
	"oam-center/models"

	"github.com/gin-gonic/gin"
)

type NginxNode struct {
	Sid    int    `json:"sid"`
	VpsId  int    `json:"vps"`
	Domain string `json:"domain"`
}

var MiniappTxt = []string{"<font color='red'>否</font>", "<font color='green'>是</font>"}

func AgentIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "平台管理",
	}
	c.HTML(200, "agentIndex", data)
}

func AgentAddIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "平台注册",
	}
	c.HTML(200, "agentAdd", data)
}

func AgentEditIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "平台编辑",
	}

	id, _ := strconv.Atoi(c.Query("id"))
	agent := models.AgentGetByAid(id)
	if agent != nil {
		data["agent"] = agent
	}
	c.HTML(200, "agentEdit", data)
}

func AgentNoticeIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "平台编辑",
	}

	id, _ := strconv.Atoi(c.Query("id"))
	agent := models.AgentGetByAid(id)
	if agent != nil {
		data["agent"] = agent
	}
	c.HTML(200, "agentNotice", data)
}

func AgentTable(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))

	agentFlag := strings.TrimSpace(c.DefaultQuery("agent_flag", ""))
	var agents []models.Agent
	var amout int
	if len(agentFlag) > 0 {
		agents, amout = models.AgentGetPageList(page, limit, "flag", agentFlag, "status", 1)
	} else {
		agents, amout = models.AgentGetPageList(page, limit, "status", 1)
	}

	list := make([]gin.H, 0, len(agents))
	for _, v := range agents {
		row := make(gin.H)
		row["aid"] = v.Aid
		row["agent_flag"] = v.Flag
		row["agent_name"] = v.Name
		row["agent_lang"] = v.Lang
		row["miniapp"] = v.Miniapp
		row["miniapp_txt"] = MiniappTxt[v.Miniapp]
		row["cdn_vps"] = v.Vpsid
		row["cdn_source"] = v.Source
		row["cdn_domain"] = v.Domain
		row["create_time"] = libs.FormatTime(v.CreateTime)
		list = append(list, row)
	}
	c.JSON(200, gin.H{"code": 0, "count": amout, "data": list})
}

func AgentIsMiniapp(c *gin.Context) {
	aid, _ := strconv.Atoi(c.PostForm("aid"))
	agent := models.AgentGetByAid(aid)
	if agent != nil {
		nginxList := make(map[int]string)
		if agent.Miniapp > 0 {
			nginxs, _ := models.NginxGetList("aid", aid)
			for _, v := range nginxs {
				nginxList[v.Id] = fmt.Sprintf("# aid=%d sid=%d vps=%d", v.Aid, v.Sid, v.Vpsid)
			}
		}

		c.JSON(200, gin.H{"status": 0, "miniapp": agent.Miniapp, "nginxs": nginxList})
	} else {
		c.JSON(200, gin.H{"status": -1, "message": "该平台不存在"})
	}
}

func checkCdnVps(vpsid int) error {
	// cdn服务器是否合法
	vps := models.VpsGetById(vpsid)
	if vps == nil {
		return fmt.Errorf("该平台的cdn云服务器不存在")
	}

	// vps是否是game
	if !vps.CheckType(models.VpsCDN) {
		return fmt.Errorf("该vps非CDN类型")
	}
	return nil
}

// 平台标示不能有下划线
func checkAgentFlag(flag string) error {
	if idx := strings.Index(flag, "_"); idx >= 0 {
		return fmt.Errorf("平台标示不能有下划线")
	}
	return nil
}

func agentUpsert(op string, c *gin.Context) (err error) {
	id, _ := strconv.Atoi(c.PostForm("aid"))
	if op == "insert" {
		// 新增
		oldAgent := models.AgentGetByAid(id)
		if oldAgent != nil {
			err = fmt.Errorf("该平台已经存在")
			return
		}

		agent := new(models.Agent)
		err = c.Bind(agent)
		if err != nil {
			return
		}

		if err = checkCdnVps(agent.Vpsid); err != nil {
			return
		}
		if err = checkAgentFlag(agent.Flag); err != nil {
			return
		}

		agent.Status = 1
		agent.CreateTime = time.Now().Unix()
		err = agent.Insert()
	} else {
		// 更新
		agent := models.AgentGetByAid(id)
		if agent == nil {
			err = fmt.Errorf("平台不存在")
			return
		} else {
			c.Bind(agent)
		}

		if err = checkCdnVps(agent.Vpsid); err != nil {
			return err
		}
		if err = checkAgentFlag(agent.Flag); err != nil {
			return err
		}
		err = agent.Update()
	}
	return err
}

func AgentAjaxSave(c *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0})
		}
	}()
	err = agentUpsert("insert", c)
}

func AgentAjaxUpdate(c *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0})
		}
	}()
	err = agentUpsert("update", c)
}

func AgentAjaxDel(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	agent := models.AgentGetByAid(id)
	if agent != nil {
		agent.Status = 0
		agent.Update()
		c.JSON(200, gin.H{"status": 0})
	}
}

// 更新服务器列表到cdn（wyftlist.json）
func AgentUpdateGsList(c *gin.Context) {
	aid, _ := strconv.Atoi(c.PostForm("id"))
	agent := models.AgentGetByAid(aid)
	if agent != nil {
		res, err := agent.UpdateGsList()
		if err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0, "message": res})
		}
	} else {
		c.JSON(200, gin.H{"status": -1, "message": "平台不存在"})
	}
}

// 更新整个平台的公告到cdn（eg.: cn_dalan/notice.json）
func AgentUpdateNotice(c *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0})
		}
	}()

	n := &models.CDNNotice{}
	err = c.Bind(n)
	if err != nil {
		return
	}

	if len(n.Title) == 0 || len(n.Content) == 0 {
		err = fmt.Errorf("请输入公告标题和公告内容")
		return
	}

	id, _ := strconv.Atoi(c.PostForm("id"))
	agent := models.AgentGetByAid(id)
	if agent != nil {
		err = agent.UpdateNotice(n, "")
	} else {
		err = fmt.Errorf("平台不存在")
	}
}

func AgentUpdateAudit(c *gin.Context) {
	aid, _ := strconv.Atoi(c.PostForm("id"))
	audit, _ := strconv.Atoi(c.PostForm("audit"))

	agent := models.AgentGetByAid(aid)
	if agent != nil && audit > 0 {
		agent.Auditversion = audit
		agent.Update("audit_version")
		c.JSON(200, gin.H{"status": 0})
	} else {
		c.JSON(200, gin.H{"status": -1, "message": "平台不存在或提审版本错误"})
	}
}

func AgentPassAudit(c *gin.Context) {
	aid, _ := strconv.Atoi(c.PostForm("id"))
	agent := models.AgentGetByAid(aid)
	if agent != nil {
		if agent.Auditversion == 0 {
			c.JSON(200, gin.H{"status": -1, "message": "非小程序，无需提审"})
			return
		}

		res, err := agent.UpdateGsList()
		if err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			agent.LastAudit = agent.Auditversion
			agent.Auditversion = 0
			agent.Update("audit_version", "last_audit")
			c.JSON(200, gin.H{"status": 0, "message": res})
		}
	} else {
		c.JSON(200, gin.H{"status": -1, "message": "平台不存在"})
	}
}
