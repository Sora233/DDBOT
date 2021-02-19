package lsp

import (
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Sora233/Sora233-MiraiGo/lsp/backend/middleware/requestid"
	"github.com/gin-gonic/gin"
)

func (*Lsp) StartBackend() {
	r := gin.Default()
	r.Use(requestid.RequestID())

	auth := r.Group("/", gin.BasicAuth(config.GlobalConfig.GetStringMapString("backend.basicAuth")))

	v1 := auth.Group("/v1")
	v1.GET("/ping", PingHandler)

	{
		v1Groups := v1.Group("/groups")
		v1Groups.GET("", GroupsGetHandler)
		v1Groups.DELETE("", GroupsDeleteHandler)
	}

	go r.Run(config.GlobalConfig.GetString("backend.addr"))
}
