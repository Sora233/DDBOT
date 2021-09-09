package concern

type Identity struct {
	id   interface{}
	name string
}

func (i *Identity) GetUid() interface{} {
	return i.id
}

func (i *Identity) GetName() string {
	return i.name
}

func NewIdentity(id interface{}, name string) *Identity {
	return &Identity{
		id:   id,
		name: name,
	}
}
