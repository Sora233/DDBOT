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
	TargetCode() int64
}

type PrivateTarget struct {
	Uin int64 `json:"uin"`
}

func (t *PrivateTarget) TargetType() TargetType {
	return TargetPrivate
}

func (t *PrivateTarget) TargetCode() int64 {
	return t.Uin
}

type GroupTarget struct {
	GroupCode int64 `json:"group_code"`
}

func (t *GroupTarget) TargetType() TargetType {
	return TargetGroup
}

func (t *GroupTarget) TargetCode() int64 {
	return t.GroupCode
}

func NewGroupTarget(groupCode int64) *GroupTarget {
	return &GroupTarget{GroupCode: groupCode}
}

func NewPrivateTarget(uin int64) *PrivateTarget {
	return &PrivateTarget{Uin: uin}
}
