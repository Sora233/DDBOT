package concern

type Identity struct {
	id    interface{}
	name  string
	type1 Type
}

func (i *Identity) Id() interface{} {
	return i.id
}

func (i *Identity) Name() string {
	return i.name
}

func (i *Identity) Type() Type {
	return i.type1
}

func NewIdentity(id interface{}, name string, ctype Type) *Identity {
	return &Identity{
		id:    id,
		name:  name,
		type1: ctype,
	}
}
