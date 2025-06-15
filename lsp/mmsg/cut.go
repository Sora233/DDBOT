package mmsg

import (
	"github.com/LagrangeDev/LagrangeGo/message"
)

type CutElement struct {
}

func (c *CutElement) Type() message.ElementType {
	return Cut
}

func (c *CutElement) PackToElement(Target) message.IMessageElement {
	return nil
}
