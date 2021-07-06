package models

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

type Cross struct {
	Id         int    `form:"id"`
	Plat       int    `db:"plat" form:"cross_plat" binding:"required"`
	Version    int    `db:"version" form:"cross_version" binding:"required"`
	Vpsid      int    `db:"vpsid" form:"cross_vps" binding:"required"`
	Port       int    `db:"port" form:"cross_port" binding:"required"`
	DbPort     int    `db:"db_port" form:"db_port" binding:"required"`
	CreateTime int64  `db:"create_time"`
	Status     int    `db:"status"`
	Procs      string `db:"procs"`
	InstallLog string `db:"install_log"`
	HotLog     string `db:"hot_log"`
	Hoted      string `db:"hoted"`
	StartLog   string `db:"start_log"`
	StopLog    string `db:"stop_log"`
}

var CrossFields = []string{
	"id",
	"plat",
	"version",
	"vpsid",
	"port",
	"db_port",
	"create_time",
	"procs",
	"status",
	"install_log",
	"hoted",
	"hot_log",
	"start_log",
	"stop_log",
}

func (c *Cross) Insert() error {
	flds := CrossFields
	vals := make([]string, 0, len(flds))
	for _, v := range flds {
		vals = append(vals, ":"+v)
	}

	sql := fmt.Sprintf("INSERT INTO kgo_cross (%s) VALUES (%s)", strings.Join(flds, ","), strings.Join(vals, ","))
	_, err := mysqlDB.NamedExec(sql, c)
	if err != nil {
		log.Println("Cross Insert Failed, err:", err)
	}
	return err
}

func (c *Cross) Update(fields ...string) error {
	if len(fields) == 0 {
		fields = CrossFields
	}

	if len(fields) > 0 {
		flds := make([]string, 0, len(fields))
		for _, v := range fields {
			flds = append(flds, fmt.Sprintf("%s=:%s", v, v))
		}

		f := strings.Join(flds, ",")
		sql := fmt.Sprintf("UPDATE kgo_cross SET %s WHERE id=:id", f)
		_, err := mysqlDB.NamedExec(sql, c)
		if err != nil {
			log.Println("Cross Update Failed, err:", err)
		}
		return err
	}
	return fmt.Errorf("Cross Update Need fields")
}

func (c *Cross) GetFlag() string {
	agent := AgentGetByAid(c.Plat)
	if agent != nil {
		return fmt.Sprintf("%s_C%d", agent.GetFlag(), c.Id)
	}
	return ""
}

func (c *Cross) GetVersionTag() string {
	agent := AgentGetByAid(c.Plat)
	if agent == nil {
		return ""
	} else {
		tag := fmt.Sprintf("%s_%d", agent.GetFlag(), c.Version)
		return tag
	}
}

func (c *Cross) genConfig(flag string) gin.H {
	agent := AgentGetByAid(c.Plat)

	// conf
	conf := make(gin.H)
	conf["debug"] = c.Id > 10000
	conf["type"] = "cross"
	conf["plat"] = c.Plat
	conf["platName"] = agent.Flag
	conf["name"] = flag
	conf["host"] = "127.0.0.1"
	conf["pport"] = c.Port
	conf["dbHost"] = "127.0.0.1"
	conf["dbPort"] = fmt.Sprintf("%d", c.DbPort)
	conf["dbUser"] = "root"
	conf["dbPass"] = "kgogame2018"
	conf["domain"] = fmt.Sprintf("c%d-%s-jmxy.kgogame.com", c.Id, agent.Flag)
	conf["openTime"] = c.CreateTime

	cgs, _ := CrossGetGames(c.Id)
	games := make([]gin.H, 0, len(cgs))
	for _, game := range cgs {
		row := gin.H{
			"group":  c.Id,
			"plat":   game.Aid,
			"server": game.Sid,
			"ip":     game.GetIp(),
			"port":   game.Port,
		}

		// 在同一台物理机上
		if game.Vpsid == c.Vpsid {
			row["ip"] = "127.0.0.1"
		}
		games = append(games, row)
	}
	conf["cross"] = games
	return conf
}

func (c *Cross) Install() bool {
	crossFlag := c.GetFlag()
	data := gin.H{
		"flag":    crossFlag,
		"version": c.Version,
		"conf":    c.genConfig(crossFlag),
	}

	ok := false
	rsp, err := HttpPostJsonByVps(c.Vpsid, "cross/install", data)
	c.Status = 0
	if err == nil {
		// 安装成功
		if rsp.Status == 0 {
			ok = true
			if len(rsp.Data) > 0 {
				m := make(map[string]string)
				json.Unmarshal(rsp.Data, &m)
				c.HotLog = m["hoted"]
			}
			c.Status = 1
		}
		c.InstallLog = rsp.Message
	} else {
		// 安装失败
		c.InstallLog = "跨服安装失败：\n" + err.Error()
	}
	c.Update()
	return ok
}

func (c *Cross) UpdateConfig() (bool, string) {
	crossFlag := c.GetFlag()
	data := gin.H{
		"flag": crossFlag,
		"conf": c.genConfig(crossFlag),
	}

	rsp, err := HttpPostJsonByVps(c.Vpsid, "cross/config", data)
	if err == nil {
		return rsp.Status == 0, rsp.Message
	} else {
		return false, err.Error()
	}
}

func (c *Cross) HotPatch() bool {
	data := gin.H{
		"flag":    c.GetFlag(),
		"version": c.Version,
	}

	ok := false
	rsp, err := HttpPostJsonByVps(c.Vpsid, "cross/hot", data)
	if err == nil {
		if len(rsp.Message) > 0 {
			c.HotLog = rsp.Message
		}
		if len(rsp.Data) > 0 {
			m := make(map[string]string)
			json.Unmarshal(rsp.Data, &m)
			c.Hoted = m["hoted"]
		}
		if rsp.Status == 0 {
			ok = true
		}
	} else {
		c.HotLog = "补丁安装失败：\n" + err.Error()
	}
	c.Update("hot_log", "hoted")
	return ok
}

func (c *Cross) UpdateVersion(version int) bool {
	data := gin.H{
		"flag":    c.GetFlag(),
		"version": version,
	}

	ok := false
	rsp, err := HttpPostJsonByVps(c.Vpsid, "cross/update", data)
	if err == nil {
		if rsp.Status == 0 && len(rsp.Data) > 0 {
			m := make(map[string]string)
			json.Unmarshal(rsp.Data, &m)
			c.Hoted = m["hoted"]
			c.Version = version
			ok = true
		}
		c.InstallLog = rsp.Message
	} else {
		c.InstallLog = "游服更新失败：\n" + err.Error()
	}
	c.Update()
	return ok
}

func (c *Cross) Start() bool {
	data := gin.H{"flag": c.GetFlag()}

	ok := false
	rsp, err := HttpPostJsonByVps(c.Vpsid, "cross/start", data)
	if err == nil {
		if len(rsp.Message) > 0 {
			c.StartLog = rsp.Message
		}
		if rsp.Status == 0 {
			ok = true
			c.Status = 2
		}
	} else {
		c.StartLog = "开服失败：\n" + err.Error()
	}
	c.Update("start_log", "status")
	return ok
}

func (c *Cross) Stop() bool {
	data := gin.H{"flag": c.GetFlag()}

	ok := false
	rsp, err := HttpPostJsonByVps(c.Vpsid, "cross/stop", data)
	if err == nil {
		if len(rsp.Message) > 0 {
			c.StopLog = rsp.Message
		}
		if rsp.Status == 0 {
			ok = true
			c.Status = 3
		}
	} else {
		c.StopLog = "关服失败：\n" + err.Error()
	}
	c.Update("stop_log", "status")
	return ok
}

func CrossGetList(filters ...interface{}) ([]Cross, int) {
	crosss := []Cross{}
	count := MysqlGetList("kgo_cross", &crosss, filters...)
	if count == 0 {
		count = len(crosss)
	}

	// 注意： for range 的坑
	/*
		for i := 0; i < len(crosss); i++ {
			v := &crosss[i]
			if len(v.GameJson) > 0 {
				err := json.Unmarshal([]byte(v.GameJson), &v.Games)
				if err != nil {
					log.Println("CrossGetList Failed, err:", err)
				}
			}
		}
	*/
	return crosss, count
}

// 分页查找
func CrossGetPageList(page, limit int, filters ...interface{}) ([]Cross, int) {
	fs := []interface{}{"_page", page, "_limit", limit}
	fs = append(fs, filters...)
	return CrossGetList(fs...)
}

func CrossGetById(id int) *Cross {
	cross := &Cross{}
	err := mysqlDB.Get(cross, "SELECT * FROM kgo_cross WHERE id=?", id)
	if err != nil {
		log.Println("CrossGetById failed, err:", err, id)
		return nil
	}
	return cross
}

// 获取跨服的游服列表
func CrossGetGames(id int) ([]Game, int) {
	return GameGetList("cid", id)
}

func CrossGenId(isDebug bool) int {
	var id int
	var err error
	if isDebug {
		err = mysqlDB.Get(&id, "SELECT MAX(id) FROM kgo_cross WHERE id>10000")
	} else {
		err = mysqlDB.Get(&id, "SELECT MAX(id) FROM kgo_cross WHERE id<=10000")
	}

	if err != nil {
		log.Println("CrossGenId failed, err:", err, id)
		if isDebug {
			return 10001
		}
		return 1
	}
	return id + 1
}
