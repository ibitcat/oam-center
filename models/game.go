package models

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"oam-center/conf"

	"github.com/gin-gonic/gin"
)

type Game struct {
	Id          int    `form:"id"`
	Aid         int    `form:"game_agent" binding:"required"`
	Sid         int    `form:"game_sid" binding:"required"`
	Serial      int    `form:"game_serial"`
	Gid         int    `db:"gid"`
	Mid         int    `db:"mid"`
	Version     int    `form:"game_version" binding:"required"`
	Name        string `db:"name" form:"game_name" binding:"required"`
	Vpsid       int    `form:"game_vps" binding:"required"`
	CreateTime  int64  `db:"create_time"`
	InstallTime int64  `db:"install_time"`
	UpdateTime  int64  `db:"update_time"`
	Port        int    `form:"game_port" binding:"required"`
	DbPort      int    `db:"db_port" form:"db_port"`
	DbShare     int    `db:"db_share"`
	OpenTime    int64  `db:"open_time"`
	MergeTime   int64  `db:"merge_time"`
	IsTls       int    `db:"is_tls" form:"is_tls"`
	Domain      string
	Procs       string
	Status      int
	NginxId     int `db:"nginx_id" form:"game_nginx"`
	Ws          string
	Single      string
	Mode        int    `form:"game_mode" binding:"required"`
	InstallLog  string `db:"install_log"`
	HotLog      string `db:"hot_log"`
	Hoted       string `db:"hoted"`
	StartLog    string `db:"start_log"`
	StopLog     string `db:"stop_log"`
	Cid         int    `db:"cid"`
	JoinTime    int64  `db:"join_time"`
}

var GameFields = []string{
	"aid",
	"sid",
	"serial",
	"gid",
	"mid",
	"version",
	"name",
	"vpsid",
	"create_time",
	"install_time",
	"update_time",
	"port",
	"db_port",
	"db_share",
	"open_time",
	"merge_time",
	"is_tls",
	"domain",
	"procs",
	"status",
	"nginx_id",
	"ws",
	"single",
	"mode",
	"install_log",
	"hoted",
	"hot_log",
	"start_log",
	"stop_log",
	"cid",
	"join_time",
}

func (g *Game) Insert() error {
	flds := GameFields

	vals := make([]string, 0, len(flds))
	for _, v := range flds {
		vals = append(vals, ":"+v)
	}

	sql := fmt.Sprintf("INSERT INTO kgo_game (%s) VALUES (%s)", strings.Join(flds, ","), strings.Join(vals, ","))
	res, err := mysqlDB.NamedExec(sql, g)
	if err == nil {
		lastId, _ := res.LastInsertId()
		g.Id = int(lastId)
	} else {
		log.Println("Game Insert Failed, err:", err)
	}
	return err
}

func (g *Game) Update(fields ...string) error {
	if len(fields) == 0 {
		fields = GameFields
	}

	if len(fields) > 0 {
		flds := make([]string, 0, len(fields))
		for _, v := range fields {
			flds = append(flds, fmt.Sprintf("%s=:%s", v, v))
		}

		f := strings.Join(flds, ",")
		sql := fmt.Sprintf("UPDATE kgo_game SET	%s WHERE id=:id", f)
		_, err := mysqlDB.NamedExec(sql, g)
		if err != nil {
			log.Println("Game Update Failed, err:", err)
		}
		return err
	}
	return fmt.Errorf("Game Update Need fields")
}

func (g *Game) GetFlag() string {
	agent := AgentGetByAid(g.Aid)
	if agent == nil {
		return ""
	} else {
		flag := fmt.Sprintf("%s_S%d", agent.GetFlag(), g.Sid)
		return flag
	}
}

func (g *Game) GetIp() string {
	vps := VpsGetById(g.Vpsid)
	if vps == nil {
		return ""
	} else {
		return vps.Ip
	}
}

func (g *Game) GetVersionTag() string {
	agent := AgentGetByAid(g.Aid)
	if agent == nil {
		return ""
	} else {
		tag := fmt.Sprintf("%s_%d", agent.GetFlag(), g.Version)
		return tag
	}
}

func (g *Game) GetMode() string {
	oamCnf := conf.YamlConf.Oam
	if g.Sid <= oamCnf.TestSid {
		return "正式"
	} else if g.Sid > oamCnf.AuditSid {
		return "提审"
	} else {
		return "外测"
	}
}

func (g *Game) Install() bool {
	agent := AgentGetByAid(g.Aid)

	// conf
	conf := make(gin.H)
	gameFlag := g.GetFlag()
	conf["debug"] = g.Sid > 10000
	conf["type"] = "game"
	conf["plat"] = g.Aid
	conf["platName"] = agent.Flag
	conf["name"] = gameFlag
	conf["host"] = "127.0.0.1"
	conf["port"] = g.Port
	conf["key"] = "GoOdManAbC"
	conf["dbHost"] = "0.0.0.0"
	conf["dbPort"] = fmt.Sprintf("%d", g.DbPort)
	conf["dbUser"] = "root"
	conf["dbPass"] = "kgogame2018"
	conf["domain"] = g.Domain
	conf["isTls"] = g.IsTls
	conf["miniapp"] = agent.Miniapp
	conf["server"] = []gin.H{gin.H{"id": g.Sid, "serial": g.Serial, "openTime": g.OpenTime}}
	conf["zip"] = "134.175.149.43"
	conf["zport"] = 9999
	if g.DbShare > 0 {
		shareGame := GameGetById(g.DbShare)
		conf["dbShare"] = shareGame.GetFlag()
	}

	data := gin.H{
		"flag":    gameFlag,
		"version": g.Version,
		"conf":    conf,
	}

	ok := false
	rsp, err := HttpPostJsonByVps(g.Vpsid, "game/install", data)
	//log.Println(rsp, err)
	if err == nil {
		// 安装成功
		if rsp.Status == 0 {
			ok = true
			if len(rsp.Data) > 0 {
				m := make(map[string]string)
				json.Unmarshal(rsp.Data, &m)
				g.HotLog = m["hoted"]
			}
		}
		g.InstallLog = rsp.Message
	} else {
		// 安装失败
		g.InstallLog = "游服安装失败：\n" + err.Error()
	}

	g.Status = 0
	if ok {
		err = g.MiniappNginx()
		if err != nil {
			g.InstallLog += ("小程序转发配置安装：\n" + err.Error())
			ok = false
		} else {
			g.Status = 1
		}
	}
	g.Update()
	return ok
}

// 小程序需要配置nginx转发
func (g *Game) MiniappNginx() error {
	if g.NginxId == 0 {
		return nil
	}

	data := gin.H{
		"gameFlag":   g.GetFlag(),
		"gameDomain": g.Domain,
		"gameIp":     g.GetIp(),
		"gamePort":   g.Port,
	}

	var err error
	var rsp *HttpJsonRsp
	nginx := NginxGetById(g.NginxId)
	rsp, err = HttpPostJsonByVps(nginx.Vpsid, "nginx/miniapp", data)
	if err == nil && rsp.Status != 0 {
		err = fmt.Errorf(rsp.Message)
	}
	return err
}

func (g *Game) UpdateVersion(version int) bool {
	data := gin.H{
		"flag":    g.GetFlag(),
		"version": version,
	}

	ok := false
	rsp, err := HttpPostJsonByVps(g.Vpsid, "game/update", data)
	if err == nil {
		if rsp.Status == 0 && len(rsp.Data) > 0 {
			m := make(map[string]string)
			json.Unmarshal(rsp.Data, &m)
			g.Hoted = m["hoted"]
			g.Version = version
			ok = true
		}
		g.InstallLog = rsp.Message
	} else {
		g.InstallLog = "游服更新失败：\n" + err.Error()
	}
	g.Update()
	return ok
}

func (g *Game) HotPatch() bool {
	data := gin.H{
		"flag":    g.GetFlag(),
		"version": g.Version,
	}

	ok := false
	rsp, err := HttpPostJsonByVps(g.Vpsid, "game/hot", data)
	if err == nil {
		if len(rsp.Message) > 0 {
			g.HotLog = rsp.Message
		}
		if len(rsp.Data) > 0 {
			m := make(map[string]string)
			json.Unmarshal(rsp.Data, &m)
			g.Hoted = m["hoted"]
		}
		if rsp.Status == 0 {
			ok = true
		}
	} else {
		g.HotLog = "补丁安装失败：\n" + err.Error()
	}
	g.Update("hot_log", "hoted")
	return ok
}

func (g *Game) Start() bool {
	data := gin.H{"flag": g.GetFlag()}

	ok := false
	rsp, err := HttpPostJsonByVps(g.Vpsid, "game/start", data)
	if err == nil {
		if len(rsp.Message) > 0 {
			g.StartLog = rsp.Message
		}
		if rsp.Status == 0 {
			ok = true
			g.Status = 2
		}
	} else {
		g.StartLog = "开服失败：\n" + err.Error()
	}
	g.Update("start_log", "status")
	return ok
}

func (g *Game) Stop() bool {
	data := gin.H{"flag": g.GetFlag()}

	ok := false
	rsp, err := HttpPostJsonByVps(g.Vpsid, "game/stop", data)
	if err == nil {
		if len(rsp.Message) > 0 {
			g.StopLog = rsp.Message
		}
		if rsp.Status == 0 {
			ok = true
			g.Status = 3
		}
	} else {
		g.StopLog = "关服失败：\n" + err.Error()
	}
	g.Update("stop_log", "status")
	return ok
}

func GameGetList(filters ...interface{}) ([]Game, int) {
	games := []Game{}
	count := MysqlGetList("kgo_game", &games, filters...)
	if count == 0 {
		count = len(games)
	}
	return games, count
}

// 分页查找
func GameGetPageList(page, limit int, filters ...interface{}) ([]Game, int) {
	fs := []interface{}{"_page", page, "_limit", limit}
	fs = append(fs, filters...)
	return GameGetList(fs...)
}

func GameGetById(id int) *Game {
	game := &Game{}
	err := mysqlDB.Get(game, "SELECT * FROM kgo_game WHERE id=?", id)
	if err != nil {
		log.Println("GameGetById failed, err:", err, id)
		return nil
	}
	return game
}

// flag = cn_andou_S1
func GameGetByFlag(flag string) *Game {
	gameRegexp := regexp.MustCompile(`\w+?_(\w+?)_S(\d+)`)
	s := gameRegexp.FindStringSubmatch(flag)
	if len(s) == 3 {
		sid, _ := strconv.Atoi(s[2])
		agent := AgentGetByFlag(s[1])
		if agent != nil {
			game := &Game{}
			err := mysqlDB.Get(game, "SELECT * FROM kgo_game WHERE aid=? AND sid=?", agent.Aid, sid)
			if err != nil {
				log.Println("GameGetByFlag failed, err:", err, flag)
				return nil
			}
			return game
		}
	}
	return nil
}

func GameGetCount() int {
	var count int
	err := mysqlDB.Get(&count, "SELECT count(*) FROM kgo_game;")
	if err != nil {
		log.Println("GameGetCount failed, err:", err.Error())
	}
	return count
}

func GameGetNearList() []Game {
	games := []Game{}
	sqlStr := "SELECT * FROM kgo_game order by create_time desc LIMIT 20"
	err := mysqlDB.Select(&games, sqlStr)
	if err != nil {
		log.Printf("GameGetSortList failed, err:%v\n", err)
		return nil
	}

	return games
}

func GameFindSerial() int {
	serials := []int{}
	err := mysqlDB.Select(&serials, "SELECT serial FROM kgo_game ORDER BY serial ASC;")
	if err != nil {
		log.Println("GameFindSerial failed, err:", err.Error())
	}

	minVal := 0
	for _, v := range serials {
		if minVal+1 < v {
			return minVal + 1
		} else {
			minVal = v
		}
	}

	if minVal+1 <= 16383 {
		return minVal + 1
	}
	return -1
}

func GameResetCid(cid int, in string) {
	var sqlStr string
	if len(in) == 0 {
		sqlStr = fmt.Sprintf("UPDATE kgo_game SET cid=0 WHERE cid=%d", cid)
	} else {
		sqlStr = fmt.Sprintf("UPDATE kgo_game SET cid=0 WHERE cid=%d AND id not in (%s)", cid, in)
	}
	mysqlDB.MustExec(sqlStr)
	//res := mysqlDB.MustExec(sqlStr)
	//log.Println(res.LastInsertId())
	//log.Println(res.RowsAffected())
}

func GameGetVersionByAid(aid int) []Game {
	game := []Game{}
	var sqlStr string
	if aid == 0 {
		sqlStr = fmt.Sprintf("SELECT version FROM kgo_game GROUP BY (version)")
	}else{
		sqlStr = fmt.Sprintf("SELECT version FROM kgo_game WHERE aid=%d GROUP BY (version)", aid)
	}
	err := mysqlDB.Select(&game, sqlStr)
	if err != nil {
		log.Println("GameGetVersionByAid failed, err:", err, aid)
		return nil
	}
	return game
}

func GameGetSidNameByAidVersion(aid int, version int,isTest int) []Game {
	game := []Game{}
	var sid string
	if isTest == 0 {
		sid = fmt.Sprintf("sid<10000")
	} else {
		sid = fmt.Sprintf("sid>=10000")
	}
	var sqlStr string
	if aid == 0 {
		sqlStr = fmt.Sprintf("SELECT * FROM kgo_game WHERE version=%d AND %s", version, sid)
	}else{
		sqlStr = fmt.Sprintf("SELECT * FROM kgo_game WHERE aid=%d AND version=%d AND %s", aid, version, sid)
	}
	err := mysqlDB.Select(&game, sqlStr)
	if err != nil {
		log.Println("GameGetSidNameByAidVersion failed, err:", err, aid)
		return nil
	}
	return game
}