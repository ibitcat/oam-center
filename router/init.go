package router

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func createMultiTpl(r *gin.Engine) {
	render := multitemplate.NewRenderer()
	render.AddFromFiles("login", "views/login/login.html")
	render.AddFromFiles("home", "views/public/main.html")
	render.AddFromFiles("start", "views/public/layout.html", "views/home/start.html")

	// 角色管理
	render.AddFromFiles("roleIndex", "views/public/layout.html", "views/role/list.html")
	render.AddFromFiles("roleEdit", "views/public/layout.html", "views/role/edit.html")
	render.AddFromFiles("roleAdd", "views/public/layout.html", "views/role/add.html")

	// 用户管理
	render.AddFromFiles("userIndex", "views/public/layout.html", "views/user/list.html")
	render.AddFromFiles("userAdd", "views/public/layout.html", "views/user/add.html")
	render.AddFromFiles("userEdit", "views/public/layout.html", "views/user/edit.html")

	// 菜单管理
	render.AddFromFiles("menuIndex", "views/public/layout.html", "views/menu/list.html")

	// 个人中心
	render.AddFromFiles("personal", "views/public/layout.html", "views/personal/edit.html")

	// 平台管理
	render.AddFromFiles("agentIndex", "views/public/layout.html", "views/agent/list.html")
	render.AddFromFiles("agentAdd", "views/public/layout.html", "views/agent/add.html")
	render.AddFromFiles("agentEdit", "views/public/layout.html", "views/agent/edit.html")
	render.AddFromFiles("agentNotice", "views/public/layout.html", "views/agent/notice.html")

	// 转发nginx管理
	render.AddFromFiles("nginxIndex", "views/public/layout.html", "views/nginx/list.html")
	render.AddFromFiles("nginxAdd", "views/public/layout.html", "views/nginx/add.html")

	// vps列表
	render.AddFromFiles("vpsIndex", "views/public/layout.html", "views/vps/list.html")

	// 游服管理
	render.AddFromFiles("gameIndex", "views/public/layout.html", "views/game/list.html")
	render.AddFromFiles("gameAdd", "views/public/layout.html", "views/game/add.html")
	render.AddFromFiles("gameEdit", "views/public/layout.html", "views/game/edit.html")
	render.AddFromFiles("gameDetail", "views/public/layout.html", "views/game/detail.html")
	render.AddFromFiles("gameBatchcontrol", "views/public/layout.html", "views/game/batchcontrol.html")

	// 版本列表
	render.AddFromFiles("versionIndex", "views/public/layout.html", "views/version/list.html")
	render.AddFromFiles("versionUpload", "views/public/layout.html", "views/version/upload.html")

	// cdn管理
	render.AddFromFiles("cdnIndex", "views/public/layout.html", "views/cdn/list.html")
	render.AddFromFiles("cdnNotice", "views/public/layout.html", "views/cdn/notice.html")

	// 跨服管理
	render.AddFromFiles("crossIndex", "views/public/layout.html", "views/cross/list.html")
	render.AddFromFiles("crossAdd", "views/public/layout.html", "views/cross/add.html")
	render.AddFromFiles("crossEdit", "views/public/layout.html", "views/cross/edit.html")
	render.AddFromFiles("crossDetail", "views/public/layout.html", "views/cross/detail.html")

	r.HTMLRender = render
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("auth")
		if user == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"flashError": "没有权限"})
		} else {
			if cookie, ok := user.(string); ok {
				s := strings.Split(cookie, "|")
				if len(s) == 3 {
					c.Set("userId", s[0])
					c.Set("userName", s[1])
					c.Next()
					return
				}
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"flashError": "没有权限"})
		}
	}
}

func RegRouter(r *gin.Engine) {
	createMultiTpl(r)

	// session中间件
	store := sessions.NewCookieStore([]byte("secret"))
	store.Options(sessions.Options{
		MaxAge: 7 * 86400,
		// Path:   "/",
	})
	r.Use(sessions.Sessions("kgogame", store))

	// 静态资源
	r.Static("/static", "./static")

	// login
	r.GET("/", LoginIndex)
	r.POST("/login", Login)
	r.GET("/logout", Logout)

	// private
	private := r.Group("/home")
	private.Use(AuthRequired())
	{
		private.GET("/", HomeIndex)
		private.GET("/start", HomeStartIndex)

		// 角色管理
		private.GET("/role", RoleIndex)
		private.GET("/role/table", RoleTable)
		private.GET("/role/add", RoleAddIndex)
		private.GET("/role/edit", RoleEditIndex)
		private.POST("/role/ajaxsave", RoleAjaxSave)

		// 用户管理
		private.GET("/user", UserIndex)
		private.GET("/user/table", UserTable)
		private.GET("/user/add", UserAddIndex)
		private.GET("/user/edit", UserEditIndex)
		private.POST("/user/ajaxsave", UserAjaxSave)
		private.POST("/user/ajaxdel", UserAjaxDel)

		// 菜单管理
		private.GET("/menu", MenuIndex)
		private.POST("/menu/getnode", GetNode)
		private.POST("/menu/getnodes", GetNodes)
		private.POST("/menu/ajaxsave", MenuAjaxSave)
		private.POST("/menu/ajaxdel", MenuAjaxDel)

		// 个人中心
		private.GET("/personal", PersonalIndex)
		private.POST("/personal/ajaxsave", PersonalAjaxSave)

		// 平台管理
		private.GET("/agent", AgentIndex)
		private.GET("/agent/table", AgentTable)
		private.GET("/agent/add", AgentAddIndex)
		private.GET("/agent/edit", AgentEditIndex)
		private.GET("/agent/notice", AgentNoticeIndex)
		private.POST("/agent/isminiapp", AgentIsMiniapp)
		private.POST("/agent/ajaxsave", AgentAjaxSave)
		private.POST("/agent/ajaxupdate", AgentAjaxUpdate)
		private.POST("/agent/ajaxdel", AgentAjaxDel)
		private.POST("/agent/ajaxgslist", AgentUpdateGsList)
		private.POST("/agent/ajaxnotice", AgentUpdateNotice)
		private.POST("/agent/updateaudit", AgentUpdateAudit)
		private.POST("/agent/passaudit", AgentPassAudit)

		// nginx管理
		private.GET("/nginx", NginxIndex)
		private.GET("/nginx/add", NginxAddIndex)
		private.GET("/nginx/table", NginxTable)
		private.POST("/nginx/ajaxsave", NginxAjaxSave)
		private.POST("/nginx/updatedomain", NginxUpdateDomain)
		private.POST("/nginx/install", NginxInstall)

		// vps列表
		private.GET("/vps", VpsIndex)
		private.GET("/vps/table", VpsTable)
		private.GET("/vps/ping", VpsPing)
		private.POST("/vps/batchping", VpsBatchPing)
		private.POST("/vps/updatectl", VpsUpdateCtl)
		private.POST("/vps/changetime", VpsChangeTime)

		// 游服管理
		private.GET("/game", GameIndex)
		private.GET("/game/table", GameTable)
		private.GET("/game/add", GameAddIndex)
		private.GET("/game/edit", GameEditIndex)
		private.GET("/game/detail", GameDetailIndex)
		private.POST("/game/ajaxsave", GameAjaxSave)
		private.POST("/game/updatemode", GameAjaxUpdateMode)
		private.POST("/game/batchinstall", GameBatchInstall)
		private.POST("/game/batchhot", GameBatchHot)
		private.POST("/game/updatever", GameUpdateVersion)
		private.POST("/game/batchstart", GameBatchStart)
		private.POST("/game/batchstop", GameBatchStop)
		private.POST("/game/updatemid", GameUpdateMid)
		// 批量操作
		private.GET("/game/batchcontrol", GameBatchControl)
		private.POST("/game/getversion", GameGetVersion)
		private.POST("/game/getgamename", GameGetName)
		private.POST("/game/batchcontrolrun", GameBatchControlRun)


		// 版本列表
		private.GET("/version", VersionIndex)
		private.GET("/version/table", VersionTable)
		private.POST("/version/hot", VersionHotPatch)
		private.GET("/version/upload", VersionUpload)
		private.POST("/version/uploadfile", VersionUploadfile)
		private.POST("/version/delfile", VersionDelfile)
		private.POST("/version/makedir", VersionMakedir)

		// cdn管理
		private.GET("/cdn", CDNIndex)
		private.GET("/cdn/table", CDNTable)
		private.GET("/cdn/notice", CDNNoticeIndex)
		private.GET("/cdn/fresh", CDNFreshVersions)
		private.GET("/cdn/cleanup", CDNCleanUp)
		private.POST("/cdn/ajaxlog", CDNAjaxLog)
		private.POST("/cdn/batchinstall", CDNInstall)
		private.POST("/cdn/stop", CDNStopRemote)
		private.POST("/cdn/ajaxnotice", CDNUpdateNotice)

		// 跨服管理
		private.GET("/cross", CrossIndex)
		private.GET("/cross/table", CrossTable)
		private.GET("/cross/add", CrossAddIndex)
		private.GET("/cross/edit", CrossEditIndex)
		private.GET("/cross/detail", CrossDetailIndex)
		private.POST("/cross/ajaxsave", CrossAjaxSave)
		private.POST("/cross/batchinstall", CrossBatchInstall)
		private.POST("/cross/batchhot", CrossBatchHot)
		private.POST("/cross/updatever", CrossUpdateVersion)
		private.POST("/cross/batchstart", CrossBatchStart)
		private.POST("/cross/batchstop", CrossBatchStop)
		private.POST("/cross/batchconfig", CrossBatchConfig)
	}

	// api
	api := r.Group("/api")
	{
		api.POST("/vpsonline", VpsOnlineApi)
		api.POST("/procs", ProcsUpdateApi)
		api.POST("/notice", UpdateNoticeApi)
	}
}
