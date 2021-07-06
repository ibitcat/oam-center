package router

import (
	"fmt"
	"html/template"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"oam-center/conf"
	"oam-center/libs"
	"oam-center/models"

	"github.com/gin-gonic/gin"
)

var gameRegexp = regexp.MustCompile(`\w+?_(\w+?)_S(\d+)`)
var GameTypeTxt = map[int]string{1: "外测", 2: "正式"}
var GameModeTxt = map[int]string{1: "火爆", 2: "维护", 3: "新服", 4: "未开服"}
var GameStatusTxt = []string{
	"<font color='gray'><i class='fa fa-minus-square'></i>未安装</font>",
	"<font color='orange'><i class='fa fa-check-square'></i>已安装</font>",
	"<font color='green'><i class='fa fa-flag'></i>运行中</font>",
	"<font color='red'><i class='fa fa-times-circle'></i>已停止</font>",
}

func GameIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "游服管理",
	}
	c.HTML(200, "gameIndex", data)
}

func GameAddIndex(c *gin.Context) {
	userName := getUserName(c)

	agents, _ := models.AgentGetList("status", 1)
	agentGroup := make(map[int]string, len(agents))
	for _, v := range agents {
		agentGroup[v.Aid] = v.Flag
	}

	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "游服注册",
		"agentGroup":    agentGroup,
	}
	c.HTML(200, "gameAdd", data)
}

func GameEditIndex(c *gin.Context) {
	userName := getUserName(c)

	id, _ := strconv.Atoi(c.Query("id"))
	game := models.GameGetById(id)
	agent := models.AgentGetByAid(game.Aid)
	agents, _ := models.AgentGetList("status", 1)
	agentGroup := make(map[int]string, len(agents))
	for _, v := range agents {
		agentGroup[v.Aid] = v.Flag
	}

	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "游服编辑",
		"game":          game,
		"miniapp":       agent.Miniapp,
		"agentId":       agent.Aid,
		"agentGroup":    agentGroup,
	}

	if game.DbShare > 0 {
		shareGame := models.GameGetById(game.DbShare)
		data["dbShareFlag"] = shareGame.GetFlag()
	}
	c.HTML(200, "gameEdit", data)
}

func GameDetailIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "游服详情",
	}

	id, _ := strconv.Atoi(c.Query("id"))
	game := models.GameGetById(id)
	if game != nil {
		agent := models.AgentGetByAid(game.Aid)
		vps := models.VpsGetById(game.Vpsid)
		data["game"] = game
		data["agent"] = agent
		data["vps"] = vps
		data["gameStatus"] = template.HTML(GameStatusTxt[game.Status])
		data["createTime"] = libs.FormatTime(game.CreateTime)
		data["openTime"] = libs.FormatTime(game.OpenTime)

		// hotedPatch
		s := strings.Split(game.Hoted, ",")
		data["hotedPatch"] = s

		// versionGroup
		data["versionGroup"] = GetVersions(agent.Flag, game.Version)

		// modeGroup
		data["modeGroup"] = GameModeTxt

		// shareDB
		if game.DbShare > 0 {
			shareGame := models.GameGetById(game.DbShare)
			data["dbShareFlag"] = shareGame.GetFlag()
		}

		// 转发nginx
		if game.NginxId > 0 {
			data["nginx"] = models.NginxGetById(game.NginxId)
		}
	}
	c.HTML(200, "gameDetail", data)
}

func GameTable(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))

	filters := make([]interface{}, 0)
	gameType, _ := strconv.Atoi(c.Query("group_id"))
	if gameType > 0 {
		filters = append(filters, "type", gameType)
	}

	gameFlag := c.DefaultQuery("game_flag", "")
	if len(gameFlag) > 0 {
		s := gameRegexp.FindStringSubmatch(gameFlag)
		log.Println(s)
		if len(s) != 3 {
			c.JSON(200, gin.H{"code": 0, "count": 0})
			return
		}

		filters = append(filters, "sid", s[2])
		agent := models.AgentGetByFlag(s[1])
		if agent != nil {
			filters = append(filters, "aid", agent.Aid)
		}
	}

	games, count := models.GameGetPageList(page, limit, filters...)
	list := make([]gin.H, 0, len(games))
	for _, v := range games {
		row := make(gin.H)
		row["id"] = v.Id
		row["game_sid"] = v.Sid
		row["game_type"] = v.GetMode()
		row["game_flag"] = v.GetFlag()
		row["game_name"] = v.Name
		row["game_serial"] = v.Serial
		row["game_version"] = v.Version
		row["game_port"] = v.Port
		row["game_dbport"] = v.DbPort
		row["game_domain"] = v.Domain
		row["game_status"] = GameStatusTxt[v.Status]
		row["game_mode"] = v.Mode
		row["game_mode_txt"] = GameModeTxt[v.Mode]
		list = append(list, row)
	}

	c.JSON(200, gin.H{"code": 0, "count": count, "data": list})
}

func checkGame(game *models.Game) error {
	vps := models.VpsGetById(game.Vpsid)
	if vps == nil {
		return fmt.Errorf("vps不存在")
	}

	// vps是否是game
	if !vps.CheckType(models.VpsGame) {
		return fmt.Errorf("该vps非游服类型")
	}

	exWhere := fmt.Sprintf("id!=%d", game.Id)
	// serial检查
	gs, _ := models.GameGetList("_where", exWhere, "serial", game.Serial)
	if len(gs) > 0 {
		return fmt.Errorf("serial冲突")
	}

	// 同vps游戏端口检查
	gs, _ = models.GameGetList("_where", exWhere, "vpsid", game.Vpsid, "port", game.Port)
	if len(gs) > 0 {
		return fmt.Errorf("该vps上已存在同端口的游服")
	}
	return nil
}

func pushToCenter(game *models.Game) error {
	err := models.CenterGetAgent(game)
	if err != nil {
		return err
	}

	err = models.CenterRegisterServer(game)
	return err
}

func GameAjaxSave(c *gin.Context) {
	shareDbFlag := c.DefaultPostForm("share_db", "")
	openTime := libs.ParseTime(c.PostForm("game_opentime"))

	var err error
	defer func() {
		if err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0})
		}
	}()

	var game *models.Game
	id, _ := strconv.Atoi(c.PostForm("id"))
	mid, _ := strconv.Atoi(c.PostForm("game_mid"))
	if id == 0 {
		game = new(models.Game)
		err = c.Bind(game)
	} else {
		game = models.GameGetById(id)
		if game == nil {
			err = fmt.Errorf("游服不存在")
			return
		} else {
			err = c.Bind(game)
		}
	}
	if err != nil {
		return
	}

	// 分配一个serial
	if game.Serial == 0 {
		serial := models.GameFindSerial()
		if serial < 0 {
			err = fmt.Errorf("游服序列号已分配完")
			return
		}

		game.Serial = serial
	}

	// 通用检查
	if err = checkGame(game); err != nil {
		return
	}

	// 数据库共享
	if len(shareDbFlag) > 0 {
		shareGame := models.GameGetByFlag(shareDbFlag)
		if shareGame == nil || shareGame.Vpsid != game.Vpsid || shareGame.DbShare > 0 {
			err = fmt.Errorf("该游服的数据库无法被共享")
			return
		}
		game.DbPort = shareGame.DbPort
		game.DbShare = shareGame.Id
	} else {
		if game.DbPort <= 0 {
			err = fmt.Errorf("游服数据库端口错误")
			return
		}

		// 同vps数据库端口检查
		exWhere := fmt.Sprintf("id!=%d", game.Id)
		gs, _ := models.GameGetList("_where", exWhere, "vpsid", game.Vpsid, "db_share", 0, "db_port", game.DbPort)
		if len(gs) > 0 {
			err = fmt.Errorf("该vps上已存在同数据库端口的游服")
			return
		}
	}

	// 检查小程序转发nginx服务器设置
	var nginx *models.Nginx
	agent := models.AgentGetByAid(game.Aid)
	if agent == nil {
		err = fmt.Errorf("该游服所属的平台不存在")
		return
	} else {
		if agent.Miniapp > 0 {
			nginx = models.NginxGetById(game.NginxId)
			if nginx == nil {
				err = fmt.Errorf("游服未设置转发云服务器id")
				return
			}
		}
	}

	game.OpenTime = openTime
	game.Domain = fmt.Sprintf("s%d-%s-jmxy.kgogame.com", game.Sid, agent.Flag)
	game.Mid = mid
	// cdn 服务器列表相关信息
	if agent.Miniapp > 0 {
		// 小程序特殊处理
		gameFlag := game.GetFlag()
		game.Ws = fmt.Sprintf("wss://%s/%s_server", nginx.Ws, gameFlag)
		game.Single = fmt.Sprintf("https://%s/%s_single", nginx.Single, gameFlag)
		game.IsTls = 0
	} else {
		if game.IsTls > 0 {
			game.Ws = fmt.Sprintf("wss://%s/ws", game.Domain)
			game.Single = fmt.Sprintf("https://%s", game.Domain)
		} else {
			game.Ws = fmt.Sprintf("ws://%s/ws", game.Domain)
			game.Single = fmt.Sprintf("http://%s", game.Domain)
		}
	}

	if !conf.YamlConf.App.Debug {
		// debug模式不需要注册到中央后台
		if err = pushToCenter(game); err != nil {
			return
		}
	}
	if id == 0 {
		game.Status = 0
		game.CreateTime = time.Now().Unix()
		err = game.Insert()
	} else {
		log.Println(game)
		err = game.Update()
	}
}

func GameAjaxUpdateMode(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	mode, _ := strconv.Atoi(c.PostForm("mode"))
	if mode <= 0 {
		c.JSON(200, gin.H{"status": -1, "message": "服务器显示模式错误"})
		return
	}

	game := models.GameGetById(id)
	if game == nil {
		c.JSON(200, gin.H{"status": -1, "message": "游服不存在"})
		return
	}

	game.Mode = mode
	game.UpdateTime = time.Now().Unix()
	if err := game.Update("mode", "update_time"); err != nil {
		// 去cdn更新服务器列表
		c.JSON(200, gin.H{"status": -1, "message": err.Error()})
	} else {
		agent := models.AgentGetByAid(game.Aid)
		res, err := agent.UpdateGsList()
		if err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0, "message": res})
		}
	}
}

func GameBatchInstall(c *gin.Context) {
	ids := c.PostForm("ids")
	if len(ids) > 0 {
		res := make([]gin.H, 0)

		var wg sync.WaitGroup
		s := strings.Split(ids, ",")
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			game := models.GameGetById(id)
			if game != nil {
				if game.Status > 0 {
					row := make(gin.H)
					row["id"] = id
					row["msg"] = fmt.Sprintf("游服[%d]已经安装过", id)
					res = append(res, row)
				} else {
					wg.Add(1)
					go func(game *models.Game) {
						defer wg.Done()
						ok := game.Install()
						row := make(gin.H)
						row["id"] = game.Id
						if ok {
							row["msg"] = fmt.Sprintf("游服[%d]安装成功", game.Id)
						} else {
							row["msg"] = fmt.Sprintf("游服[%d]安装失败", game.Id)
						}
						res = append(res, row)
					}(game)
				}
			}
		}

		wg.Wait()
		c.JSON(200, gin.H{"status": 0, "data": res})
		//c.JSON(200, gin.H{"status": 0, "message": strings.Join(res, "<br>")})
	} else {
		c.JSON(200, gin.H{"status": -1, "message": "请先选择要操作的游服！"})
	}
}

func GameBatchHot(c *gin.Context) {
	ids := c.PostForm("ids")
	if len(ids) > 0 {
		var wg sync.WaitGroup

		res := make([]gin.H, 0)
		s := strings.Split(ids, ",")
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			game := models.GameGetById(id)

			row := make(gin.H)
			row["id"] = game.Id
			if game != nil && game.Status > 0 {
				// 补丁列表
				patchs := GetHotZips(game.GetVersionTag())
				if len(patchs) > 0 {
					wg.Add(1)
					go func(game *models.Game) {
						defer wg.Done()
						ok := game.HotPatch()
						if ok {
							row["msg"] = fmt.Sprintf("游服[%d]安装补丁成功", game.Id)
						} else {
							row["msg"] = fmt.Sprintf("游服[%d]安装补丁失败", game.Id)
						}
					}(game)
				} else {
					row["msg"] = "无补丁安装"
				}
			} else {
				row["msg"] = fmt.Sprintf("游服[%d]无法安装补丁", id)
			}
			res = append(res, row)
		}

		wg.Wait()
		c.JSON(200, gin.H{"status": 0, "data": res})
	}
}

func GameUpdateVersion(c *gin.Context) {
	version := c.PostForm("version")
	id, _ := strconv.Atoi(c.PostForm("id"))
	if id > 0 && len(version) > 0 {
		if !isVersionExist(version) {
			c.JSON(200, gin.H{"status": -1, "message": "版本不存在"})
			return
		}

		game := models.GameGetById(id)
		if game != nil {
			var msg string
			switch game.Status {
			case 0:
				msg = "游服尚未安装，请先安装游服!"
			case 2:
				msg = "游服正在运行，请先关闭游服!"
			}

			if len(msg) > 0 {
				c.JSON(200, gin.H{"status": -1, "message": msg})
				return
			}

			s := strings.Split(version, "_")
			ver, _ := strconv.Atoi(s[2])
			if game.Version == ver {
				c.JSON(200, gin.H{"status": -1, "message": "无需更新"})
				return
			}

			ok := game.UpdateVersion(ver)
			if ok {
				if !conf.YamlConf.App.Debug {
					// debug模式不需要注册到中央后台
					if err := pushToCenter(game); err != nil {
						log.Println("game Update pushToCenter fail, err = ", err)
					}
				}
				c.JSON(200, gin.H{"status": 0})
				return
			} else {
				c.JSON(200, gin.H{"status": -1, "message": "游服更新失败"})
				return
			}
		}
	}

	c.JSON(200, gin.H{"status": -1, "message": "请先选择要更新的版本和游服！"})
}

func GameBatchStart(c *gin.Context) {
	ids := c.PostForm("ids")
	if len(ids) > 0 {
		var wg sync.WaitGroup

		res := make([]gin.H, 0)
		s := strings.Split(ids, ",")
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			game := models.GameGetById(id)

			row := make(gin.H)
			row["id"] = game.Id
			if game != nil {
				switch game.Status {
				case 0:
					row["msg"] = fmt.Sprintf("游服[%d]尚未安装", game.Id)
				case 2:
					row["msg"] = fmt.Sprintf("游服[%d]正在运行", game.Id)
				case 1, 3:
					wg.Add(1)
					go func(game *models.Game) {
						defer wg.Done()
						ok := game.Start()
						if ok {
							row["msg"] = fmt.Sprintf("游服[%d]启动成功", game.Id)
						} else {
							row["msg"] = fmt.Sprintf("游服[%d]启动失败", game.Id)
						}
					}(game)
				default:
					row["msg"] = fmt.Sprintf("游服[%d]状态错误", id)
				}
			} else {
				row["msg"] = fmt.Sprintf("游服[%d]不存在", id)
			}
			res = append(res, row)
		}

		wg.Wait()
		c.JSON(200, gin.H{"status": 0, "data": res})
	}
}

func GameBatchStop(c *gin.Context) {
	ids := c.PostForm("ids")
	if len(ids) > 0 {
		var wg sync.WaitGroup

		res := make([]gin.H, 0)
		s := strings.Split(ids, ",")
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			game := models.GameGetById(id)

			row := make(gin.H)
			row["id"] = game.Id
			if game != nil && game.Status == 2 {
				wg.Add(1)
				go func(game *models.Game) {
					defer wg.Done()
					ok := game.Stop()
					if ok {
						row["msg"] = fmt.Sprintf("游服[%d]关闭成功", game.Id)
					} else {
						row["msg"] = fmt.Sprintf("游服[%d]关闭失败", game.Id)
					}
				}(game)
			} else {
				row["msg"] = fmt.Sprintf("游服[%d]不存在或未运行中", id)
			}
			res = append(res, row)
		}

		wg.Wait()
		c.JSON(200, gin.H{"status": 0, "data": res})
	}
}

func GameBatchControl(c *gin.Context) {
	userName := getUserName(c)

	agents, _ := models.AgentGetList("status", 1)
	agentGroup := make(map[int]string, len(agents))
	for _, v := range agents {
		agentGroup[v.Aid] = v.Flag
	}

	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "批量操作",
		"agentGroup":    agentGroup,
	}
	c.HTML(200, "gameBatchcontrol", data)
}

func GameGetVersion(c *gin.Context){
	aid, _ := strconv.Atoi(c.Query("aid"))
	version := models.GameGetVersionByAid(aid)
	if version != nil{
		c.JSON(200, gin.H{"code": 0, "message": "success" , "data":version})
	}else{
		c.JSON(200, gin.H{"code": -1, "message": "fail"})
	}
	
}

func GameGetName(c *gin.Context){
	aid, _ := strconv.Atoi(c.Query("aid"))
	version, _ := strconv.Atoi(c.Query("version"))
	isTest, _ := strconv.Atoi(c.Query("isTest"))
	gamename := models.GameGetSidNameByAidVersion(aid,version,isTest)
	if gamename != nil{
		c.JSON(200, gin.H{"code": 0, "message": "success" , "data":gamename})
	}else{
		c.JSON(200, gin.H{"code": -1, "message": "fail"})
	}
	
}

func GameBatchControlRun(c *gin.Context){
	games := c.PostForm("games")
	model := c.PostForm("batch_model")
	res := make([]gin.H, 0)
	s := strings.Split(games, ",")
	var wg sync.WaitGroup
	if model == "start" {
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			game := models.GameGetById(id)

			row := make(gin.H)
			row["id"] = game.Id
			n := strings.Split(game.Domain, "-")
			if game != nil {
				switch game.Status {
				case 0:
					row["msg"] = fmt.Sprintf("游服[%s-%s]尚未安装", n[1], n[0])
				case 2:
					row["msg"] = fmt.Sprintf("游服[%s-%s]正在运行", n[1], n[0])
				case 1, 3:
					wg.Add(1)
					go func(game *models.Game) {
						defer wg.Done()
						ok := game.Start()
						if ok {
							row["msg"] = fmt.Sprintf("游服[%s-%s]启动成功", n[1], n[0])
						} else {
							row["msg"] = fmt.Sprintf("游服[%s-%s]启动失败", n[1], n[0])
						}
					}(game)
				default:
					row["msg"] = fmt.Sprintf("游服[%s-%s]状态错误", n[1], n[0])
				}
			} else {
				row["msg"] = fmt.Sprintf("游服[%s-%s]不存在", n[1], n[0])
			}
			res = append(res, row)
		}

		wg.Wait()
		c.JSON(200, gin.H{"status": 0, "data": res})
	}else if model == "stop" {
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			game := models.GameGetById(id)

			row := make(gin.H)
			row["id"] = game.Id
			n := strings.Split(game.Domain, "-")
			if game != nil && game.Status == 2 {
				wg.Add(1)
				go func(game *models.Game) {
					defer wg.Done()
					ok := game.Stop()
					if ok {
						row["msg"] = fmt.Sprintf("游服[%s-%s]关闭成功", n[1], n[0])
					} else {
						row["msg"] = fmt.Sprintf("游服[%s-%s]关闭失败", n[1], n[0])
					}
				}(game)
			} else {
				row["msg"] = fmt.Sprintf("游服[%s-%s]不存在或未运行中", n[1], n[0])
			}
			res = append(res, row)
		}

		wg.Wait()
		c.JSON(200, gin.H{"status": 0, "data": res})
	}else if model == "update" {
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			game := models.GameGetById(id)
			agent := models.AgentGetByAid(game.Aid)

			row := make(gin.H)
			row["id"] = game.Id
			n := strings.Split(game.Domain, "-")
			if game != nil {
				switch game.Status {
				case 0:
					row["msg"] = fmt.Sprintf("游服[%s-%s]尚未安装", n[1], n[0])
				case 2:
					row["msg"] = fmt.Sprintf("游服[%s-%s]正在运行", n[1], n[0])
				case 1, 3:
					versionlist := GetVersions(agent.Flag, game.Version)
					if len(versionlist) > 0 {
						for _, ver := range versionlist {
							s := strings.Split(ver, "_")
							version, _ := strconv.Atoi(s[2])
							// log.Println(version)
							ok := game.UpdateVersion(version)
							if ok {
								row["msg"] = fmt.Sprintf("游服[%s-%s]更新%d版本成功", n[1], n[0] ,version)
							} else {
								row["msg"] = fmt.Sprintf("游服[%s-%s]更新%d版本失败", n[1], n[0] ,version)
							}
						}					
					} else {
						row["msg"] = fmt.Sprintf("游服[%s-%s]无需更新", n[1], n[0])
					}
				default:
					row["msg"] = fmt.Sprintf("游服[%s-%s]状态错误", n[1], n[0])
						}
			} else {
				row["msg"] = fmt.Sprintf("游服[%s-%s]不存在", n[1], n[0])
			}
			res = append(res, row)
		}

		c.JSON(200, gin.H{"status": 0, "data": res})
	}else if model == "install" {
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			game := models.GameGetById(id)
			n := strings.Split(game.Domain, "-")
			if game != nil {
				if game.Status > 0 {
					row := make(gin.H)
					row["id"] = id
					row["msg"] = fmt.Sprintf("游服[%s-%s]已经安装过", n[1], n[0])
					res = append(res, row)
				} else {
					wg.Add(1)
					go func(game *models.Game) {
						defer wg.Done()
						ok := game.Install()
						row := make(gin.H)
						row["id"] = game.Id
						if ok {
							row["msg"] = fmt.Sprintf("游服[%s-%s]安装成功", n[1], n[0])
						} else {
							row["msg"] = fmt.Sprintf("游服[%s-%s]安装失败", n[1], n[0])
						}
						res = append(res, row)
					}(game)
				}
			}
		}

		wg.Wait()
		c.JSON(200, gin.H{"status": 0, "data": res})
	}else if model == "patch" {
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			game := models.GameGetById(id)

			row := make(gin.H)
			row["id"] = game.Id
			n := strings.Split(game.Domain, "-")
			if game != nil && game.Status > 0 {
				// 补丁列表
				patchs := GetHotZips(game.GetVersionTag())
				if len(patchs) > 0 {
					wg.Add(1)
					go func(game *models.Game) {
						defer wg.Done()
						ok := game.HotPatch()
						if ok {
							row["msg"] = fmt.Sprintf("游服[%s-%s]安装补丁成功", n[1], n[0])
						} else {
							row["msg"] = fmt.Sprintf("游服[%s-%s]安装补丁失败", n[1], n[0])
						}
					}(game)
				} else {
					row["msg"] = "无补丁安装"
				}
			} else {
				row["msg"] = fmt.Sprintf("游服[%s-%s]无法安装补丁", n[1], n[0])
			}
			res = append(res, row)
		}

		wg.Wait()
		c.JSON(200, gin.H{"status": 0, "data": res})
	}else if model == "mode" {
		mode, _ := strconv.Atoi(c.PostForm("mode"))
		var gameAid []int
		var newAid []int
		// 更新数据库
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			game := models.GameGetById(id)
			row := make(gin.H)
			row["id"] = game.Id
			gameAid = append(gameAid, game.Aid)
			game.Mode = mode
			game.UpdateTime = time.Now().Unix()
			n := strings.Split(game.Domain, "-")
			if err := game.Update("mode", "update_time"); err != nil {
				// 更新数据库
				row["msg"] = err.Error()
				res = append(res, row)
			} else {
					row["msg"] = fmt.Sprintf("游服[%s-%s]更新模式成功", n[1], n[0])
					res = append(res, row)
				}
			}
		//平台ID去重
	    for i := 0; i < len(gameAid); i++ {
	        repeat := false
	        for j := i + 1; j < len(gameAid); j++ {
	            if gameAid[i] == gameAid[j] {
	                repeat = true
	                break
	            }
	        }
	        if !repeat {
	            newAid = append(newAid, gameAid[i])
	        }
	    }

	   //更新服务器列表文件
	   for i := 0; i < len(newAid); i++ {
			agent := models.AgentGetByAid(newAid[i])
			_, err := agent.UpdateGsList()
			row := make(gin.H)
			row["id"] = newAid[i]
			if err != nil {
				row["msg"] = err.Error()
				res = append(res, row)
			} else {
				row["msg"] = fmt.Sprintf("平台[%s]更新服务器列表文件成功", agent.Flag)
				res = append(res, row)
			}
	   }
		c.JSON(200, gin.H{"status": 0, "data": res})
	}else {
		c.JSON(200, gin.H{"status": -1, "message": "fail"})
	}

	
}

func GameUpdateMid(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	mid, _ := strconv.Atoi(c.PostForm("mid"))
	game := models.GameGetById(id)
	if game == nil {
		c.JSON(200, gin.H{"status": -1, "message": "游服不存在"})
		return
	}

	game.Mid = mid
	game.UpdateTime = time.Now().Unix()
	if err := game.Update("mid", "update_time"); err != nil {
		// 更新数据库
		c.JSON(200, gin.H{"status": -1, "message": err.Error()})
	} else {
			c.JSON(200, gin.H{"status": 0})
		}
}