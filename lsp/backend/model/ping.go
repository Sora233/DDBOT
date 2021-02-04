package model

type PingResponse struct {
	GenericHeader
	Message string `json:"message"`
}
