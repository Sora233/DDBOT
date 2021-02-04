package model

import (
	"github.com/Sora233/Sora233-MiraiGo/lsp/backend/middleware/requestid"
	"github.com/gin-gonic/gin"
)

type GenericHeader struct {
	RequestID    string `json:"request_id"`
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message,omitempty"`
}

func NewGenericHeader(ctx *gin.Context, code int, err error) GenericHeader {
	gh := GenericHeader{
		RequestID: requestid.GetRequestID(ctx),
		ErrorCode: code,
	}
	if code != ErrOk && err == nil {
		err = GetError(code)
	}
	if err != nil {
		gh.ErrorMessage = err.Error()
	}
	return gh
}
