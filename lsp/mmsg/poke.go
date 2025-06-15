package mmsg

import (
	"github.com/LagrangeDev/LagrangeGo/message"

	"github.com/Sora233/MiraiGo-Template/bot"
)

// PokeElement 戳一戳
type PokeElement struct {
	Uin uint32
}

func NewPoke(uin uint32) *PokeElement {
	return &PokeElement{Uin: uin}
}

func (p *PokeElement) Type() message.ElementType {
	return Poke
}

func (p *PokeElement) PackToElement(target Target) message.IMessageElement {
	switch target.TargetType() {
	case TargetGroup:
		bot.QQClient.GroupPoke(target.TargetCode(), p.Uin)
	case TargetPrivate:
		bot.QQClient.FriendPoke(target.TargetCode())
	}
	return nil
}
