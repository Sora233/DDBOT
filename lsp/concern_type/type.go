package concern_type

import (
	"sort"
	"strings"
)

// Type 是ddbot标记的订阅类型
type Type string

// TypeSeparator 是 Type 的分割符号，所以定义 Type 时，不可以包含这个符号
const TypeSeparator = "/"

// Empty 表示一个空 Type
const Empty Type = ""

// FromString 从string中解析Type
func FromString(s string) Type {
	return Type(s)
}

// String 把Type转成string格式
func (t Type) String() string {
	return string(t)
}

// IsTrivial 检查t是否是单个type，即不是由多个type加起来的
func (t Type) IsTrivial() bool {
	return !strings.Contains(string(t), TypeSeparator)
}

// Empty 如果t是空Type，返回true
func (t Type) Empty() bool {
	return t == Empty
}

// Split 把t拆分成基本单位
func (t Type) Split() []Type {
	if t.Empty() {
		return nil
	}
	return split(t)
}

// ContainAll 如果t包含o的所有type，返回true
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

// ContainAny 如果t包含o内任意一个type，返回true
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

// Remove 从t内删除传入的type
// 注意这个操作并不会改变t，而是返回操作后的type
// 如果要改变t，可以这样： t = t.Remove(...)
func (t Type) Remove(oList ...Type) Type {
	var typeSet = make(map[Type]interface{})
	ts := t.Split()
	for _, tp := range ts {
		typeSet[tp] = struct{}{}
	}
	for _, o := range oList {
		os := o.Split()
		for _, tp := range os {
			delete(typeSet, tp)
		}
	}
	var result []Type
	for k := range typeSet {
		result = append(result, k)
	}
	return combine(result)
}

// Add 把传入的type加入到t中
// 注意这个操作并不会改变t，而是返回操作后的type
// 如果要改变t，可以这样： t = t.Add(...)
func (t Type) Add(oList ...Type) Type {
	var typeSet = make(map[Type]interface{})
	ts := t.Split()
	for _, tp := range ts {
		typeSet[tp] = struct{}{}
	}
	for _, o := range oList {
		os := o.Split()
		for _, tp := range os {
			typeSet[tp] = struct{}{}
		}
	}
	var result []Type
	for k := range typeSet {
		result = append(result, k)
	}
	return combine(result)
}

// Intersection 返回t和o的交集
// 注意这个操作并不会改变t，而是返回操作后的type
// 如果要改变t，可以这样： t = t.Intersection(...)
func (t Type) Intersection(o Type) Type {
	var s1 = make(map[Type]interface{})
	for _, u1 := range t.Split() {
		s1[u1] = true
	}
	var result []Type
	for _, u2 := range o.Split() {
		if _, found := s1[u2]; found {
			result = append(result, u2)
		}
	}
	return combine(result)
}

func combine(t []Type) Type {
	if t == nil {
		return Empty
	}
	sort.Slice(t, func(i, j int) bool {
		return t[i] < t[j]
	})
	var sb strings.Builder
	for index, r := range t {
		if index > 0 {
			sb.WriteString(TypeSeparator)
		}
		sb.WriteString(r.String())
	}
	return FromString(sb.String())
}

func split(t Type) []Type {
	spt := strings.Split(string(t), TypeSeparator)
	sort.Strings(spt)
	var result []Type
	for _, s := range spt {
		result = append(result, Type(s))
	}
	return result
}
