package pgr

import "reflect"

// And creates AND from a list of conditions.
func And(cond ...Builder) Builder {
	return BuildFunc(func(buf Buffer) error {
		return buildCond(buf, "AND", cond...)
	})
}

// Or creates OR from a list of conditions.
func Or(cond ...Builder) Builder {
	return BuildFunc(func(buf Buffer) error {
		return buildCond(buf, "OR", cond...)
	})
}

func buildCond(buf Buffer, pred string, cond ...Builder) error {
	for i, c := range cond {
		if i > 0 {
			buf.WriteString(" ")
			buf.WriteString(pred)
			buf.WriteString(" ")
		}
		buf.WriteString("(")
		err := c.Build(buf)
		if err != nil {
			return err
		}
		buf.WriteString(")")
	}
	return nil
}

func buildCmp(buf Buffer, pred string, column string, value interface{}) error {
	buf.WriteString(QuoteIdent(column))
	buf.WriteString(" ")
	buf.WriteString(pred)
	buf.WriteString(" ")
	buf.WriteString(placeholder)

	buf.WriteValue(value)
	return nil
}

// Eq is `=`.
// When value is nil, it will be translated to `IS NULL`.
// When value is a slice, it will be translated to `IN`.
// Otherwise it will be translated to `=`.
func Eq(column string, value interface{}) Builder {
	return BuildFunc(func(buf Buffer) error {
		if value == nil {
			buf.WriteString(QuoteIdent(column))
			buf.WriteString(" IS NULL")
			return nil
		}
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Slice {
			if v.Len() == 0 {
				buf.WriteString(EncodeBool(false))
				return nil
			}
			return buildCmp(buf, "IN", column, value)
		}
		return buildCmp(buf, "=", column, value)
	})
}

// Neq is `!=`.
// When value is nil, it will be translated to `IS NOT NULL`.
// When value is a slice, it will be translated to `NOT IN`.
// Otherwise it will be translated to `!=`.
func Neq(column string, value interface{}) Builder {
	return BuildFunc(func(buf Buffer) error {
		if value == nil {
			buf.WriteString(QuoteIdent(column))
			buf.WriteString(" IS NOT NULL")
			return nil
		}
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Slice {
			if v.Len() == 0 {
				buf.WriteString(EncodeBool(true))
				return nil
			}
			return buildCmp(buf, "NOT IN", column, value)
		}
		return buildCmp(buf, "!=", column, value)
	})
}

// Gt is `>`.
func Gt(column string, value interface{}) Builder {
	return BuildFunc(func(buf Buffer) error {
		return buildCmp(buf, ">", column, value)
	})
}

// Gte is '>='.
func Gte(column string, value interface{}) Builder {
	return BuildFunc(func(buf Buffer) error {
		return buildCmp(buf, ">=", column, value)
	})
}

// Lt is '<'.
func Lt(column string, value interface{}) Builder {
	return BuildFunc(func(buf Buffer) error {
		return buildCmp(buf, "<", column, value)
	})
}

// Lte is `<=`.
func Lte(column string, value interface{}) Builder {
	return BuildFunc(func(buf Buffer) error {
		return buildCmp(buf, "<=", column, value)
	})
}

func buildLike(buf Buffer, column, pattern string, isNot bool, escape []string) error {
	buf.WriteString(QuoteIdent(column))
	if isNot {
		buf.WriteString(" NOT LIKE ")
	} else {
		buf.WriteString(" LIKE ")
	}
	buf.WriteString(EncodeString(pattern))
	if len(escape) > 0 {
		buf.WriteString(" ESCAPE ")
		buf.WriteString(EncodeString(escape[0]))
	}
	return nil
}

// Like is `LIKE`, with an optional `ESCAPE` clause
func Like(column, value string, escape ...string) Builder {
	return BuildFunc(func(buf Buffer) error {
		return buildLike(buf, column, value, false, escape)
	})
}

// NotLike is `NOT LIKE`, with an optional `ESCAPE` clause
func NotLike(column, value string, escape ...string) Builder {
	return BuildFunc(func(buf Buffer) error {
		return buildLike(buf, column, value, true, escape)
	})
}

func buildIlike(buf Buffer, column, pattern string, isNot bool, escape []string) error {
	buf.WriteString(QuoteIdent(column))
	if isNot {
		buf.WriteString(" NOT ILIKE ")
	} else {
		buf.WriteString(" ILIKE ")
	}
	buf.WriteString(EncodeString(pattern))
	if len(escape) > 0 {
		buf.WriteString(" ESCAPE ")
		buf.WriteString(EncodeString(escape[0]))
	}
	return nil
}

func Ilike(column, value string, escape ...string) Builder {
	return BuildFunc(func(buf Buffer) error {
		return buildIlike(buf, column, value, false, escape)
	})
}

func NotIlike(column, value string, escape ...string) Builder {
	return BuildFunc(func(buf Buffer) error {
		return buildIlike(buf, column, value, true, escape)
	})
}
