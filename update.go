package queryx

type UpdateBuilder struct{}

func (db queryx) Update(table string) *UpdateBuilder {
	return &UpdateBuilder{}
}
