package lsp

import (
	"github.com/Sora233/Sora233-MiraiGo/lsp/backend/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

func PingHandler(c *gin.Context) {
	var (
		response model.PingResponse
		errCode  int
		err      error
	)
	defer func() {
		if e := recover(); e != nil {
			errCode = model.ErrInternal
		}
		response.Header = model.NewGenericHeader(c, errCode, err)
		c.JSON(http.StatusOK, response)
	}()
	response.Message = "pong"
}

func GroupsGetHandler(c *gin.Context) {

}

func GroupsDeleteHandler(c *gin.Context) {

}
