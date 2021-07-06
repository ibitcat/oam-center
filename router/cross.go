package router

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"sync"
	"time"

	"oam-center/libs"
	"oam-center/models"

	"github.com/gin-gonic/gin"
)

var CrossStatusTxt = []string{
	"<font color='gray'><i class='fa fa-minus-square'></i>未安装</font>",
	"<font color='orange'><i class='fa fa-check-square'></i>已安装</font>",
	"<font color='green'><i class='fa fa-flag'></i>运行中</font>",
	"<font color='red'><i class='fa fa-times-circle'></i>已停止</font>",
}

func CrossIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "跨服管理",
	}
	c.HTML(200, "crossIndex", data)
}

func CrossAddIndex(c *gin.Context) {
	userName := getUserName(c)

	agents, _ := models.AgentGetList("status", 1)
	agentGroup := make(map[int]string, len(agents))
	for _, v := range agents {
		agentGroup[v.Aid] = v.Flag
	}

	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "新建跨服",
		"agentGroup":    agentGroup,
	}
	c.HTML(200, "crossAdd", data)
}

func CrossEditIndex(c *gin.Context) {
	userName := getUserName(c)

	id, _ := strconv.Atoi(c.Query("id"))
	cross := models.CrossGetById(id)

	agents, _ := models.AgentGetList("status", 1)
	agentGroup := make(map[int]string, len(agents))
	for _, v := range agents {
		agentGroup[v.Aid] = v.Flag
	}

	cgs, count := models.CrossGetGames(id)
	games := make([]gin.H, 0, count)
	for _, g := range cgs {
		games = append(games, gin.H{"game": g.GetFlag()})
	}

	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "修改跨服",
		"agentGroup":    agentGroup,
		"cross":         cross,
		"games":         games,
	}
	c.HTML(200, "crossEdit", data)
}

func CrossDetailIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "跨服详情",
	}

	id, _ := strconv.Atoi(c.Query("id"))
	cross := models.CrossGetById(id)
	if cross != nil {
		agent := models.AgentGetByAid(cross.Plat)
		vps := models.VpsGetById(cross.Vpsid)
		data["cross"] = cross
		data["agent"] = agent
		data["vps"] = vps
		data["crossStatus"] = template.HTML(CrossStatusTxt[cross.Status])
		data["createTime"] = libs.FormatTime(cross.CreateTime)

		// hotedPatch
		s := strings.Split(cross.Hoted, ",")
		data["hotedPatch"] = s

		// versionGroup
		data["versionGroup"] = GetVersions(agent.Flag, cross.Version)
	}
	c.HTML(200, "crossDetail", data)
}

func CrossTable(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))

	filters := []interface{}{"_order", "id"}
	groupId, _ := strconv.Atoi(c.Query("group_id"))
	if groupId > 0 {
		filters = append(filters, "id", groupId)
	}

	crosss, count := models.CrossGetPageList(page, limit, filters...)
	list := make([]gin.H, 0, len(crosss))
	for _, v := range crosss {
		row := make(gin.H)
		row["id"] = v.Id
		row["cross_plat"] = v.GetFlag()
		row["cross_version"] = v.Version
		row["cross_port"] = v.Port
		row["cross_dbport"] = v.DbPort
		row["create_time"] = libs.FormatTime(v.CreateTime)

		//games
		games, _ := models.CrossGetGames(v.Id)
		flags := make([]string, 0, len(games))
		times := make([]string, 0, len(games))
		for _, g := range games {
			times = append(times, libs.FormatTime(g.JoinTime))
			flags = append(flags, fmt.Sprintf(`<a href="/home/game/detail?id=%d" class="news-item-title" target="_blank">%s # %d</a>`, g.Id, g.GetFlag(), g.Id))
		}
		row["game_flag"] = template.HTML(strings.Join(flags, "<br>"))
		row["join_time"] = template.HTML(strings.Join(times, "<br>"))
		row["cross_status"] = CrossStatusTxt[v.Status]
		list = append(list, row)
	}

	c.JSON(200, gin.H{"code": 0, "count": count, "data": list})
}

func checkCross(cross *models.Cross) error {
	// vps检查
	vps := models.VpsGetById(cross.Vpsid)
	if vps == nil {
		return fmt.Errorf("vps不存在")
	}
	if !vps.CheckType(models.VpsCross) {
		return fmt.Errorf("该vps非跨服类型")
	}

	where := fmt.Sprintf("id!=%d", cross.Id)
	// 同vps游戏端口检查
	cs, _ := models.CrossGetList("vpsid", cross.Vpsid, "port", cross.Port, "_where", where)
	if len(cs) > 0 {
		return fmt.Errorf("该vps上已存在同端口的跨服")
	}

	// 同vps跨服数据库端口检查
	cs, _ = models.CrossGetList("vpsid", cross.Vpsid, "db_port", cross.DbPort, "_where", where)
	if len(cs) > 0 {
		return fmt.Errorf("该vps上已存在同数据库端口的跨服")
	}

	return nil
}

func CrossAjaxSave(c *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			c.JSON(200, gin.H{"status": -1, "message": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": 0})
		}
	}()

	ctm := time.Now().Unix()
	var cross *models.Cross

	isDebug := c.PostForm("debug") == "on"
	id, _ := strconv.Atoi(c.PostForm("id"))
	if id == 0 {
		gId := models.CrossGenId(isDebug)
		cross = new(models.Cross)
		err = c.Bind(cross)

		cross.Id = gId
		cross.Status = 0
		cross.CreateTime = ctm
	} else {
		cross = models.CrossGetById(id)
		if cross == nil {
			err = fmt.Errorf("跨服不存在")
			return
		} else {
			err = c.Bind(cross)
		}
	}
	if err != nil {
		return
	}

	// 通用检查
	if err = checkCross(cross); err != nil {
		return
	}

	// 游服列表
	gamesTxt := c.PostForm("games")
	flags := strings.Split(gamesTxt, ",")
	games := make([]*models.Game, 0, len(flags))
	if len(gamesTxt) > 0 {
		for _, v := range flags {
			game := models.GameGetByFlag(v)
			if game != nil && game.Aid == cross.Plat {
				// 是否安装
				if game.Status == 0 {
					err = fmt.Errorf("游服[%s]未安装", v)
					return
				}

				// 是否已经跨服
				if game.Cid > 0 && game.Cid != cross.Id {
					err = fmt.Errorf("游服[%s]已经加入其它跨服", v)
					return
				}
				games = append(games, game)
			} else {
				err = fmt.Errorf("游服标示[%s]错误或平台不匹配", v)
				return
			}
		}
	}

	//if !conf.YamlConf.App.Debug {
	// debug模式不需要注册到中央后台
	// TODO
	//}

	if id == 0 {
		err = cross.Insert()
	} else {
		err = cross.Update()
	}

	if err == nil {
		sqlIn := make([]string, 0, len(flags))
		for _, g := range games {
			g.Cid = cross.Id
			if g.JoinTime == 0 {
				g.JoinTime = ctm
			}
			sqlIn = append(sqlIn, fmt.Sprintf("%d", g.Id))
			g.Update("cid", "join_time")
		}
		models.GameResetCid(cross.Id, strings.Join(sqlIn, ","))
	}
}

func CrossBatchInstall(c *gin.Context) {
	ids := c.PostForm("ids")
	if len(ids) > 0 {
		res := make([]gin.H, 0)

		var wg sync.WaitGroup
		s := strings.Split(ids, ",")
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			cross := models.CrossGetById(id)
			if cross != nil {
				if cross.Status > 0 {
					row := make(gin.H)
					row["id"] = id
					row["msg"] = fmt.Sprintf("跨服[%d]已经安装过", id)
					res = append(res, row)
				} else {
					wg.Add(1)
					go func(cross *models.Cross) {
						defer wg.Done()
						ok := cross.Install()
						row := make(gin.H)
						row["id"] = cross.Id
						if ok {
							row["msg"] = fmt.Sprintf("跨服[%d]安装成功", cross.Id)
						} else {
							row["msg"] = fmt.Sprintf("跨服[%d]安装失败", cross.Id)
						}
						res = append(res, row)
					}(cross)
				}
			}
		}

		wg.Wait()
		c.JSON(200, gin.H{"status": 0, "data": res})
	} else {
		c.JSON(200, gin.H{"status": -1, "message": "请先选择要操作的跨服！"})
	}
}

func CrossBatchHot(c *gin.Context) {
	ids := c.PostForm("ids")
	if len(ids) > 0 {
		var wg sync.WaitGroup

		res := make([]gin.H, 0)
		s := strings.Split(ids, ",")
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			cross := models.CrossGetById(id)

			row := make(gin.H)
			row["id"] = cross.Id
			if cross != nil && cross.Status > 0 {
				// 补丁列表
				patchs := GetHotZips(cross.GetVersionTag())
				if len(patchs) > 0 {
					wg.Add(1)
					go func(cross *models.Cross) {
						defer wg.Done()
						ok := cross.HotPatch()
						if ok {
							row["msg"] = fmt.Sprintf("跨服[%d]安装补丁成功", cross.Id)
						} else {
							row["msg"] = fmt.Sprintf("跨服[%d]安装补丁失败", cross.Id)
						}
					}(cross)
				} else {
					row["msg"] = "无补丁安装"
				}
			} else {
				row["msg"] = fmt.Sprintf("跨服[%d]无法安装补丁", id)
			}
			res = append(res, row)
		}

		wg.Wait()
		c.JSON(200, gin.H{"status": 0, "data": res})
	}
}

func CrossUpdateVersion(c *gin.Context) {
	version := c.PostForm("version")
	id, _ := strconv.Atoi(c.PostForm("id"))
	if id > 0 && len(version) > 0 {
		if !isVersionExist(version) {
			c.JSON(200, gin.H{"status": -1, "message": "版本不存在"})
			return
		}

		cross := models.CrossGetById(id)
		if cross != nil {
			var msg string
			switch cross.Status {
			case 0:
				msg = "跨服尚未安装，请先安装跨服!"
			case 2:
				msg = "跨服正在运行，请先关闭跨服!"
			}

			if len(msg) > 0 {
				c.JSON(200, gin.H{"status": -1, "message": msg})
				return
			}

			s := strings.Split(version, "_")
			ver, _ := strconv.Atoi(s[2])
			if cross.Version == ver {
				c.JSON(200, gin.H{"status": -1, "message": "无需更新"})
				return
			}

			ok := cross.UpdateVersion(ver)
			if ok {
				c.JSON(200, gin.H{"status": 0})
				return
			} else {
				c.JSON(200, gin.H{"status": -1, "message": "跨服更新失败"})
				return
			}
		}
	}

	c.JSON(200, gin.H{"status": -1, "message": "请先选择要更新的版本和跨服！"})
}

func CrossBatchStart(c *gin.Context) {
	ids := c.PostForm("ids")
	if len(ids) > 0 {
		var wg sync.WaitGroup

		res := make([]gin.H, 0)
		s := strings.Split(ids, ",")
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			cross := models.CrossGetById(id)

			row := make(gin.H)
			row["id"] = cross.Id
			if cross != nil {
				switch cross.Status {
				case 0:
					row["msg"] = fmt.Sprintf("跨服[%d]尚未安装", cross.Id)
				case 2:
					row["msg"] = fmt.Sprintf("跨服[%d]正在运行", cross.Id)
				case 1, 3:
					wg.Add(1)
					go func(cross *models.Cross) {
						defer wg.Done()
						ok := cross.Start()
						if ok {
							row["msg"] = fmt.Sprintf("跨服[%d]启动成功", cross.Id)
						} else {
							row["msg"] = fmt.Sprintf("跨服[%d]启动失败", cross.Id)
						}
					}(cross)
				default:
					row["msg"] = fmt.Sprintf("跨服[%d]状态错误", id)
				}
			} else {
				row["msg"] = fmt.Sprintf("跨服[%d]不存在", id)
			}
			res = append(res, row)
		}

		wg.Wait()
		c.JSON(200, gin.H{"status": 0, "data": res})
	}
}

func CrossBatchStop(c *gin.Context) {
	ids := c.PostForm("ids")
	if len(ids) > 0 {
		var wg sync.WaitGroup

		res := make([]gin.H, 0)
		s := strings.Split(ids, ",")
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			cross := models.CrossGetById(id)

			row := make(gin.H)
			row["id"] = cross.Id
			if cross != nil && cross.Status == 2 {
				wg.Add(1)
				go func(cross *models.Cross) {
					defer wg.Done()
					ok := cross.Stop()
					if ok {
						row["msg"] = fmt.Sprintf("跨服[%d]关闭成功", cross.Id)
					} else {
						row["msg"] = fmt.Sprintf("跨服[%d]关闭失败", cross.Id)
					}
				}(cross)
			} else {
				row["msg"] = fmt.Sprintf("跨服[%d]不存在或未运行中", id)
			}
			res = append(res, row)
		}

		wg.Wait()
		c.JSON(200, gin.H{"status": 0, "data": res})
	}
}

func CrossBatchConfig(c *gin.Context) {
	ids := c.PostForm("ids")
	if len(ids) > 0 {
		var wg sync.WaitGroup

		res := make([]gin.H, 0)
		s := strings.Split(ids, ",")
		for _, v := range s {
			id, _ := strconv.Atoi(v)
			cross := models.CrossGetById(id)

			row := make(gin.H)
			row["id"] = cross.Id
			if cross != nil && cross.Status > 0 {
				wg.Add(1)
				go func(cross *models.Cross) {
					defer wg.Done()
					ok, res := cross.UpdateConfig()
					if ok {
						row["msg"] = fmt.Sprintf("更新跨服[%d]配置成功", cross.Id)
					} else {
						row["msg"] = fmt.Sprintf("更新跨服[%d]配置失败,err=%s", cross.Id, res)
					}
				}(cross)
			} else {
				row["msg"] = fmt.Sprintf("跨服[%d]不存在或未安装", id)
			}
			res = append(res, row)
		}

		wg.Wait()
		c.JSON(200, gin.H{"status": 0, "data": res})
	}
}
