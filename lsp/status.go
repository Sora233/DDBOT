package lsp

type Status struct {
	ImagePoolEnable bool
	ProxyPoolEnable bool
}

func NewStatus() *Status {
	c := &Status{}
	return c
}
