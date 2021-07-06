package router

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func getClientIp(c *gin.Context) string {
	s := strings.Split(c.Request.RemoteAddr, ":")
	return s[0]
}

func getUserId(c *gin.Context) int {
	if v, ok := c.Get("userId"); ok {
		userId, _ := strconv.Atoi(v.(string))
		return userId
	}
	return 0
}

func getUserName(c *gin.Context) string {
	if v, ok := c.Get("userName"); ok {
		name := v.(string)
		return name
	}
	return ""
}
