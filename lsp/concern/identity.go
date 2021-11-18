package concern

// Identity 是一个 IdentityInfo 的默认实现
type Identity struct {
	id   interface{}
	name string
}

// GetUid 返回 uid
func (i *Identity) GetUid() interface{} {
	return i.id
}

// GetName 返回 name
func (i *Identity) GetName() string {
	return i.name
}

// NewIdentity 根据 id 和 name 创建新的 Identity
func NewIdentity(id interface{}, name string) *Identity {
	return &Identity{
		id:   id,
		name: name,
	}
}
