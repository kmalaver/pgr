package queryx

type SelectBuilder struct{}

func (db *queryx) Select(cols ...string) *SelectBuilder {
	return &SelectBuilder{}
}
