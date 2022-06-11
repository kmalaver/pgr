package queryx

import (
	"context"
	"strconv"
)

type UpdateBuilder struct {
	runner
	raw

	Table        string
	Value        map[string]interface{}
	WhereCond    []Builder
	ReturnColumn []string
	LimitCount   int64
}

func (db queryx) Update(table string) *UpdateBuilder {
	return &UpdateBuilder{
		Table:      table,
		Value:      make(map[string]interface{}),
		LimitCount: -1,
		runner:     db.conn,
	}
}

func (db queryx) UpdateSql(query string, value ...interface{}) *UpdateBuilder {
	return &UpdateBuilder{
		raw: raw{
			Query: query,
			Value: value,
		},
		Value:      make(map[string]interface{}),
		LimitCount: -1,
		runner:     db.conn,
	}
}

// Where adds a where condition.
// query can be Builder or string. value is used only if query type is string.
func (b *UpdateBuilder) Where(query interface{}, value ...interface{}) *UpdateBuilder {
	switch query := query.(type) {
	case string:
		b.WhereCond = append(b.WhereCond, Expr(query, value...))
	case Builder:
		b.WhereCond = append(b.WhereCond, query)
	}
	return b
}

// Returning specifies the returning columns for postgres.
func (b *UpdateBuilder) Returning(column ...string) *UpdateBuilder {
	b.ReturnColumn = column
	return b
}

// Set updates column with value.
func (b *UpdateBuilder) Set(column string, value interface{}) *UpdateBuilder {
	b.Value[column] = value
	return b
}

// SetMap specifies a map of (column, value) to update in bulk.
func (b *UpdateBuilder) SetMap(m map[string]interface{}) *UpdateBuilder {
	for col, val := range m {
		b.Set(col, val)
	}
	return b
}

// IncrBy increases column by value
func (b *UpdateBuilder) IncrBy(column string, value interface{}) *UpdateBuilder {
	b.Value[column] = Expr("? + ?", I(column), value)
	return b
}

// DecrBy decreases column by value
func (b *UpdateBuilder) DecrBy(column string, value interface{}) *UpdateBuilder {
	b.Value[column] = Expr("? - ?", I(column), value)
	return b
}

func (b *UpdateBuilder) Limit(n uint64) *UpdateBuilder {
	b.LimitCount = int64(n)
	return b
}

func (b *UpdateBuilder) Exec(ctx context.Context) (int64, error) {
	res, err := exec(ctx, b.runner, b)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (b *UpdateBuilder) Load(ctx context.Context, value interface{}) error {
	_, err := query(ctx, b.runner, b, value)
	return err
}

func (b *UpdateBuilder) Build(buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.Build(buf)
	}

	if b.Table == "" {
		return ErrTableNotSpecified
	}

	if len(b.Value) == 0 {
		return ErrColumnNotSpecified
	}

	buf.WriteString("UPDATE ")
	buf.WriteString(QuoteIdent(b.Table))
	buf.WriteString(" SET ")

	i := 0
	for col, v := range b.Value {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(QuoteIdent(col))
		buf.WriteString(" = ")
		buf.WriteString(placeholder)

		buf.WriteValue(v)

		i++
	}

	if len(b.WhereCond) > 0 {
		buf.WriteString(" WHERE ")
		err := And(b.WhereCond...).Build(buf)
		if err != nil {
			return err
		}
	}

	if len(b.ReturnColumn) > 0 {
		buf.WriteString(" RETURNING ")
		for i, col := range b.ReturnColumn {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(QuoteIdent(col))
		}
	}

	if b.LimitCount >= 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.FormatInt(b.LimitCount, 10))
	}

	return nil
}
