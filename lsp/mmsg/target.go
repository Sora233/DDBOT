package mmsg

type TargetType int32

const (
	TargetGroup TargetType = iota
	TargetPrivate
)

func (t TargetType) IsGroup() bool {
	return t == TargetGroup
}

func (t TargetType) IsPrivate() bool {
	return t == TargetPrivate
}

type Target interface {
	TargetType() TargetType
	TargetCode() uint32
}

type PrivateTarget struct {
	Uin uint32 `json:"uin"`
}

func (t *PrivateTarget) TargetType() TargetType {
	return TargetPrivate
}

func (t *PrivateTarget) TargetCode() uint32 {
	return t.Uin
}

type GroupTarget struct {
	GroupCode uint32 `json:"group_code"`
}

func (t *GroupTarget) TargetType() TargetType {
	return TargetGroup
}

func (t *GroupTarget) TargetCode() uint32 {
	return t.GroupCode
}

func NewGroupTarget(groupCode uint32) *GroupTarget {
	return &GroupTarget{GroupCode: groupCode}
}

func NewPrivateTarget(uin uint32) *PrivateTarget {
	return &PrivateTarget{Uin: uin}
}
