package queryx

type DeleteBuilder struct{}

func (db queryx) DeleteFrom(table string) *DeleteBuilder {
	return &DeleteBuilder{}
}
