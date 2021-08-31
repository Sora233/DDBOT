package concern_manager

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/sirupsen/logrus"
	"strconv"
)

type Type int64

const Empty Type = 0

type Notify interface {
	Type() Type
	GetGroupCode() int64
	GetUid() interface{}
	ToMessage() []message.IMessageElement
	Logger() *logrus.Entry
}

func (t Type) String() string {
	return strconv.FormatInt(int64(t), 10)
}

func (t Type) ContainAll(o Type) bool {
	if t.Empty() && o.Empty() {
		return false
	}
	return t&o == o
}

func (t Type) ContainAny(o Type) bool {
	if t.Empty() && o.Empty() {
		return false
	}
	for !o.Empty() {
		if t&(o&(-o)) != 0 {
			return true
		}
		o -= o & (-o)
	}
	return false
}

// Split return a Type unit slice from a given Type
func (t Type) Split() []Type {
	var result []Type
	for t > 0 {
		result = append(result, t&(-t))
		t -= t & (-t)
	}
	return result
}

func (t Type) Remove(o Type) Type {
	newT := t
	for o > 0 {
		if t&(o&(-o)) != 0 {
			newT ^= o & (-o)
		}
		o -= o & (-o)
	}
	return newT
}

func (t Type) Add(o Type) Type {
	return t | o
}

func (t Type) Empty() bool {
	return t == 0
}

func FromString(s string) Type {
	t, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return Type(0)
	}
	return Type(t)
}

type Concern interface {
	Describe(ctype concern.Type)
}
