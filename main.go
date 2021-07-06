package main

import (
	"log"

	"oam-center/conf"
	"oam-center/models"
	"oam-center/router"

	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	models.Init()

	r := gin.Default()
	router.RegRouter(r)
	appYaml := conf.YamlConf.App
	if appYaml.Debug {
		r.Run(":" + appYaml.Port)
	} else {
		r.Run("127.0.0.1:" + appYaml.Port)
	}
}
