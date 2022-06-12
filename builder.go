package pgr

type Builder interface {
	Build(Buffer) error
}

// BuildFunc implements Builder.
type BuildFunc func(Buffer) error

// Build calls itself to build SQL.
func (b BuildFunc) Build(buf Buffer) error {
	return b(buf)
}
