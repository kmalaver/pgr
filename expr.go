package pgr

type raw struct {
	Query string
	Value []interface{}
}

// Expr allows raw expression to be used when current SQL syntax is not supported.
func Expr(query string, value ...interface{}) Builder {
	return &raw{Query: query, Value: value}
}

func (raw *raw) Build(buf Buffer) error {
	buf.WriteString(raw.Query)
	buf.WriteValue(raw.Value...)
	return nil
}
