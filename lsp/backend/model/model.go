package model

import (
	"github.com/Sora233/Sora233-MiraiGo/lsp/backend/middleware/requestid"
	"github.com/gin-gonic/gin"
)

func NewGenericHeader(ctx *gin.Context, code int, err error) *GenericHeader {
	gh := &GenericHeader{
		RequestId: requestid.GetRequestID(ctx),
		ErrorCode: int32(code),
	}
	if code != ErrOk && err == nil {
		err = GetError(code)
	}
	if err != nil {
		gh.ErrorMessage = err.Error()
	}
	return gh
}
