package mmsg

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
)

type CutElement struct {
}

func (c *CutElement) Type() message.ElementType {
	return Cut
}

func (c *CutElement) PackToElement(mt.Target) message.IMessageElement {
	return nil
}
