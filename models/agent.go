package models

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"oam-center/conf"
)

type GsNode map[string]string
type CDNGsList struct {
	Server       []GsNode `json:"server"`                 // 正式服务器列表
	TestServer   []GsNode `json:"testServer,omitempty"`   // 测试服务器列表
	AuditServer  []GsNode `json:"auditServer,omitempty"`  // 提审服务器列表（针对小程序）
	AuditVersion string   `json:"auditVersion,omitempty"` // 提审版本（eg.: v20190101）
}
type CDNNotice struct {
	Title   string `json:"title" form:"title" binding:"required"`
	Content string `json:"content" form:"content" binding:"required"`
	Debug   int    `json:"debug" form:"debug"`
	Version string `json:"version" form:"version"`
}

type Agent struct {
	Aid          int    `db:"aid" form:"aid" binding:"required"`
	Flag         string `db:"flag" form:"agent_flag" binding:"required"`
	Name         string `db:"name" form:"agent_name" binding:"required"`
	Lang         string `db:"lang" form:"agent_lang" binding:"required"`
	Miniapp      int    `db:"miniapp" form:"miniapp"`
	Auditversion int    `db:"audit_version"`
	LastAudit    int    `db:"last_audit"`
	Vpsid        int    `db:"vpsid" form:"vpsid" binding:"required"`
	Source       string `db:"source" form:"source" binding:"required"`
	Domain       string `db:"domain" form:"domain" binding:"required"`
	CreateTime   int64  `db:"create_time"`
	Status       int
}

var AgentFields = []string{
	"aid",
	"flag",
	"name",
	"lang",
	"miniapp",
	"audit_version",
	"last_audit",
	"vpsid",
	"source",
	"domain",
	"create_time",
	"status",
}

func (a *Agent) Insert() error {
	flds := AgentFields

	vals := make([]string, 0, len(flds))
	for _, v := range flds {
		vals = append(vals, ":"+v)
	}

	sql := fmt.Sprintf("INSERT INTO kgo_agent (%s) VALUES (%s)", strings.Join(flds, ","), strings.Join(vals, ","))
	_, err := mysqlDB.NamedExec(sql, a)
	if err != nil {
		log.Println("Agent Insert Failed, err:", err)
	}
	return err
}

func (a *Agent) Update(fields ...string) error {
	if len(fields) == 0 {
		fields = AgentFields
	}

	if len(fields) > 0 {
		flds := make([]string, 0, len(fields))
		for _, v := range fields {
			flds = append(flds, fmt.Sprintf("%s=:%s", v, v))
		}

		f := strings.Join(flds, ",")
		sql := fmt.Sprintf("UPDATE kgo_agent SET %s WHERE aid=:aid", f)
		_, err := mysqlDB.NamedExec(sql, a)
		if err != nil {
			log.Println("Agent Update Failed, err:", err)
		}
		return err
	}
	return fmt.Errorf("Agent Update Need fields")
}

func (a *Agent) UpdateGsList() (string, error) {
	// 找到该平台下的所有服务器
	games, _ := GameGetList("aid", a.Aid)

	auditVer := ""
	if a.Miniapp > 0 && a.Auditversion > 0 {
		auditVer = fmt.Sprintf("v%d", a.Auditversion)
	}
	serves := make([]GsNode, 0, len(games))
	tests := make([]GsNode, 0)
	audits := make([]GsNode, 0)

	for _, game := range games {
		node := make(GsNode)
		if game.Mid == 0 {
			node["id"] = fmt.Sprintf("%d", game.Sid)
		} else {
			node["id"] = fmt.Sprintf("%d", game.Mid)
		}		
		node["name"] = game.Name
		node["status"] = fmt.Sprintf("%d", game.Mode)
		node["addr"] = game.Ws
		node["api"] = game.Single



		oamCnf := conf.YamlConf.Oam
		if game.Sid <= oamCnf.TestSid {
			// 正式服
			serves = append(serves, node)
		} else if game.Sid > oamCnf.AuditSid {
			// 提审服
			audits = append(audits, node)
		} else {
			// 测试服
			tests = append(tests, node)
		}
	}

	gss := CDNGsList{
		Server:       serves,
		TestServer:   tests,
		AuditServer:  audits,
		AuditVersion: auditVer,
	}

	//fmt.Printf("%#v\n", gss)
	data := map[string]interface{}{
		"agent":  a.Lang + "_" + a.Flag,
		"domain": a.Domain,
		"data":   gss,
	}
	rsp, err := HttpPostJsonByVps(a.Vpsid, "cdn/gslist", data)
	if err != nil {
		return "", err
	} else {
		return rsp.Message, nil
	}
}

func (a *Agent) GetFlag() string {
	return fmt.Sprintf("%s_%s", a.Lang, a.Flag)
}

func (a *Agent) UpdateNotice(n *CDNNotice, version string) error {
	n.Content = strings.Replace(n.Content, "\r\n", `<br>`, -1)
	n.Content = strings.Replace(n.Content, "    ", "&nbsp;&nbsp;&nbsp;&nbsp;", -1)
	notice, err := JSONMarshal(n)
	if err != nil {
		return err
	}

	isDebug := n.Debug > 0
	data := map[string]interface{}{
		"debug":   isDebug,
		"agent":   a.Lang + "_" + a.Flag,
		"data":    string(notice),
		"version": version,
		"domain":  a.Domain,
	}

	rsp, err := HttpPostJsonByVps(a.Vpsid, "cdn/notice", data)
	if err == nil {
		if rsp.Status != 0 {
			err = fmt.Errorf(rsp.Message)
		}
	}
	return err
}

func (a *Agent) StopCDN(version string) error {
	data := url.Values{}
	data.Set("tag", version)

	rsp, err := HttpPostFormByVps(a.Vpsid, "cdn/stop", data)
	log.Println(rsp)
	if err == nil {
		if rsp.Status != 0 {
			err = fmt.Errorf(rsp.Message)
		}
	}
	return err
}

func AgentGetList(filters ...interface{}) ([]Agent, int) {
	agents := []Agent{}
	count := MysqlGetList("kgo_agent", &agents, filters...)
	if count == 0 {
		count = len(agents)
	}
	return agents, count
}

// 分页查找
func AgentGetPageList(page, limit int, filters ...interface{}) ([]Agent, int) {
	fs := []interface{}{"_page", page, "_limit", limit}
	fs = append(fs, filters...)
	return AgentGetList(fs...)
}

func AgentGetByAid(aid int) *Agent {
	agent := &Agent{}
	err := mysqlDB.Get(agent, "SELECT * FROM kgo_agent WHERE aid=?", aid)
	if err != nil {
		log.Println("AgentGetByAid failed, err:", err, aid)
		return nil
	}
	return agent
}

func AgentGetByFlag(flag string) *Agent {
	agent := &Agent{}
	err := mysqlDB.Get(agent, "SELECT * FROM kgo_agent WHERE flag=?", flag)
	if err != nil {
		log.Println("AgentGetByFlag failed, err:", err, flag)
		return nil
	}
	return agent
}

func AgentGetCount() int {
	var count int
	err := mysqlDB.Get(&count, "SELECT count(*) FROM kgo_agent;")
	if err != nil {
		log.Println("AgentGetCount failed, err:", err.Error())
	}
	return count
}
