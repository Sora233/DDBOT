package model

import (
	"errors"
	"fmt"
)

var (
	ErrOk            = 0
	ErrInternal      = 1001
	ErrInvalidParams = 1002
)

var ErrMsgMap = map[int]string{
	ErrOk:            "",
	ErrInternal:      "Internal Error",
	ErrInvalidParams: "Invalid Params",
}

func GetError(code int) error {
	return errors.New(GetErrorMessage(code))
}

func GetErrorMessage(code int) string {
	if msg, found := ErrMsgMap[code]; found {
		return msg
	}
	return fmt.Sprintf("unknown error code: %v", code)

}
