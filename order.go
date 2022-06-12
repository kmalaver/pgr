package pgr

type direction bool

const (
	asc  direction = false
	desc           = true
)

func order(column string, dir direction) Builder {
	return BuildFunc(func(buf Buffer) error {
		buf.WriteString(column)
		switch dir {
		case asc:
			buf.WriteString(" ASC")
		case desc:
			buf.WriteString(" DESC")
		}
		return nil
	})
}
