package concern

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
)

type Type string

const TypeSeparator = "/"

const Empty Type = ""

type Notify interface {
	Type() Type
	GetGroupCode() int64
	GetUid() interface{}
	ToMessage() []message.IMessageElement
	Logger() *logrus.Entry
}

func (t Type) String() string {
	return string(t)
}

func (t Type) IsTrivial() bool {
	return !strings.Contains(string(t), TypeSeparator)
}

func (t Type) Empty() bool {
	return t == Empty
}

// Split return a Type unit slice from a given Type
func (t Type) Split() []Type {
	if t.Empty() {
		return nil
	}
	var result []Type
	spt := strings.Split(string(t), TypeSeparator)
	for _, s := range spt {
		result = append(result, Type(s))
	}
	return result
}

func (t Type) ContainAll(o Type) bool {
	if t.Empty() && o.Empty() {
		return false
	}
	if o.Empty() {
		return true
	}
	ts := t.Split()
	os := o.Split()
	for _, u := range os {
		var ok = false
		for _, v := range ts {
			if u == v {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

func (t Type) ContainAny(o Type) bool {
	if t.Empty() && o.Empty() {
		return false
	}
	if o.Empty() {
		return true
	}
	ts := t.Split()
	os := o.Split()
	for _, u := range os {
		for _, v := range ts {
			if u == v {
				return true
			}
		}
	}
	return false
}

func (t Type) Remove(o Type) Type {
	var typeSet = make(map[Type]interface{})
	ts := t.Split()
	os := o.Split()
	for _, tp := range ts {
		typeSet[tp] = struct{}{}
	}
	for _, tp := range os {
		delete(typeSet, tp)
	}
	var result []Type
	for k := range typeSet {
		result = append(result, k)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

func (t Type) Add(o Type) Type {
	var typeSet = make(map[Type]interface{})
	ts := t.Split()
	os := o.Split()
	for _, s := range [][]Type{ts, os} {
		for _, tp := range s {
			typeSet[tp] = struct{}{}
		}
	}
	var result []Type
	for k := range typeSet {
		result = append(result, k)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

func FromString(s string) Type {
	return Type(s)
}

type Concern interface {
	Name() string
	Start() error
	Stop()
	ParseId(string) (interface{}, error)

	GetStateManager() IStateManager
	FreshIndex(groupCode ...int64)
}
