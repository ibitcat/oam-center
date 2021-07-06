package router

import (
	"strconv"
	"strings"
	"time"

	"oam-center/libs"
	"oam-center/models"

	"github.com/gin-gonic/gin"
)

var RoleStatusText = map[int]string{
	0: "<font color='red'>禁用</font>",
	1: "正常",
}

func RoleIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"pageTitle":     "角色管理",
	}
	c.HTML(200, "roleIndex", data)
}

func RoleAddIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"zTree":         true,
		"pageTitle":     "新增角色",
	}
	c.HTML(200, "roleAdd", data)
}

func RoleEditIndex(c *gin.Context) {
	userName := getUserName(c)
	data := gin.H{
		"siteName":      SiteName,
		"loginUserName": userName,
		"zTree":         true,
		"pageTitle":     "编辑角色",
	}

	// 角色信息
	id, _ := strconv.Atoi(c.Query("id"))
	role := models.RoleGetById(id)
	if role != nil {
		row := make(map[string]interface{})
		row["id"] = role.Id
		row["role_name"] = role.RoleName
		row["detail"] = role.Detail
		data["role"] = row

		// 角色的权限列表
		auths := models.MenuGetAuth(role.Id)
		authId := make([]int, 0, len(auths))
		for _, v := range auths {
			authId = append(authId, v.Id)
		}
		data["auth"] = authId
	}

	c.HTML(200, "roleEdit", data)
}

func RoleAjaxSave(c *gin.Context) {
	roleId, _ := strconv.Atoi(c.PostForm("id"))
	auths := strings.TrimSpace(c.DefaultPostForm("nodes_data", ""))

	userId := getUserId(c)
	if roleId == 0 {
		// 新增
		role := new(models.Role)
		c.Bind(role)

		role.CreateTime = time.Now().Unix()
		role.UpdateTime = time.Now().Unix()
		role.CreateId = userId
		role.UpdateId = userId
		role.Status = 1

		if err := role.Insert(); err != nil {
			c.JSON(200, gin.H{"status": -1, "message": "新增角色失败"})
		} else {
			roleId = role.Id

			// 更新菜单权限
			authsSlice := strings.Split(auths, ",")
			for _, v := range authsSlice {
				aid, _ := strconv.Atoi(v)
				menu := models.MenuGetById(aid)
				if menu != nil {
					menu.AuthBit |= (1 << uint(roleId-1))
					menu.Update()
				}
			}
			c.JSON(200, gin.H{"status": 0})
		}
	} else {
		role := models.RoleGetById(roleId)
		if role == nil {
			c.JSON(200, gin.H{"status": -1, "message": "角色不存在"})
		} else {
			c.Bind(role)
		}

		// 修改
		role.UpdateId = userId
		role.UpdateTime = time.Now().Unix()
		role.UpdateId = userId
		if err := role.Update("role_name", "detail", "update_id", "update_time"); err != nil {
			c.JSON(200, gin.H{"status": -1, "message": "更新角色失败"})
		} else {
			authsSlice := strings.Split(auths, ",")
			menus := models.MenuGetAuth(roleId)
			for _, menu := range menus {
				hasDel := true
				for _, v := range authsSlice {
					aid, _ := strconv.Atoi(v)
					if aid == menu.Id {
						hasDel = false
						break
					}
				}

				if hasDel {
					menu.AuthBit &= ^(1 << uint(roleId-1))
					menu.Update()
				}
			}

			// 新增的权限
			for _, v := range authsSlice {
				aid, _ := strconv.Atoi(v)
				menu := models.MenuGetById(aid)
				if menu != nil && menu.Status > 0 {
					menu.AuthBit |= (1 << uint(roleId-1))
					menu.Update()
				}
			}

			c.JSON(200, gin.H{"status": 0})
		}
	}
}

func RoleAjaxDel(c *gin.Context) {
	roleId, _ := strconv.Atoi(c.PostForm("id"))
	role := models.RoleGetById(roleId)
	if role != nil {
		role.Status = 0
		role.UpdateId = getUserId(c)
		role.UpdateTime = time.Now().Unix()
		role.Update("status", "update_id", "update_time")

		// 角色对应的菜单权限要取消
		menus := models.MenuGetAuth(roleId)
		for _, menu := range menus {
			menu.AuthBit &= ^(1 << uint(roleId-1))
			menu.Update()
		}
	}
}

func RoleTable(c *gin.Context) {
	//列表
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))
	roleName := strings.TrimSpace(c.DefaultQuery("roleName", ""))

	var count int
	var roles []models.Role
	if len(roleName) > 0 {
		roles, count = models.RoleGetPageList(page, limit, "role_name", roleName)
	} else {
		roles, count = models.RoleGetPageList(page, limit)
	}

	list := make([]gin.H, 0, len(roles))
	for _, v := range roles {
		row := make(gin.H)
		row["id"] = v.Id
		row["role_name"] = v.RoleName
		row["detail"] = v.Detail
		row["create_time"] = libs.FormatTime(v.CreateTime)
		row["update_time"] = libs.FormatTime(v.UpdateTime)
		row["status_text"] = RoleStatusText[v.Status]
		list = append(list, row)
	}
	c.JSON(200, gin.H{"code": 0, "count": count, "data": list})
}
