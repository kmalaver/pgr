package pgr

import (
	"context"
)

type UpdateBuilder struct {
	db *Pgr
	raw

	table        string
	value        map[string]interface{}
	whereCond    []Builder
	returnColumn []string
}

func (db *Pgr) Update(table string) *UpdateBuilder {
	return &UpdateBuilder{
		table: table,
		value: make(map[string]interface{}),
		db:    db,
	}
}

func (db *Pgr) UpdateSql(query string, value ...interface{}) *UpdateBuilder {
	return &UpdateBuilder{
		raw: raw{
			Query: query,
			Value: value,
		},
		value: make(map[string]interface{}),
		db:    db,
	}
}

// Where adds a where condition.
// query can be Builder or string. value is used only if query type is string.
func (b *UpdateBuilder) Where(query interface{}, value ...interface{}) *UpdateBuilder {
	switch query := query.(type) {
	case string:
		b.whereCond = append(b.whereCond, Expr(query, value...))
	case Builder:
		b.whereCond = append(b.whereCond, query)
	}
	return b
}

// Returning specifies the returning columns for postgres.
func (b *UpdateBuilder) Returning(column ...string) *UpdateBuilder {
	b.returnColumn = column
	return b
}

// Set updates column with value.
func (b *UpdateBuilder) Set(column string, value interface{}) *UpdateBuilder {
	b.value[column] = value
	return b
}

// SetMap specifies a map of (column, value) to update in bulk.
func (b *UpdateBuilder) SetMap(m map[string]interface{}) *UpdateBuilder {
	for col, val := range m {
		b.Set(col, val)
	}
	return b
}

func (b *UpdateBuilder) Exec(ctx context.Context) (int64, error) {
	return b.db.exec(ctx, b)
}

func (b *UpdateBuilder) Load(ctx context.Context, value interface{}) error {
	_, err := b.db.query(ctx, b, value)
	return err
}

func (b *UpdateBuilder) Build(buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.Build(buf)
	}

	if b.table == "" {
		return ErrTableNotSpecified
	}

	if len(b.value) == 0 {
		return ErrColumnNotSpecified
	}

	buf.WriteString("UPDATE ")
	buf.WriteString(QuoteIdent(b.table))
	buf.WriteString(" SET ")

	i := 0
	for col, v := range b.value {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(QuoteIdent(col))
		buf.WriteString(" = ")
		buf.WriteString(placeholder)

		buf.WriteValue(v)

		i++
	}

	if len(b.whereCond) > 0 {
		buf.WriteString(" WHERE ")
		err := And(b.whereCond...).Build(buf)
		if err != nil {
			return err
		}
	}

	if len(b.returnColumn) > 0 {
		buf.WriteString(" RETURNING ")
		for i, col := range b.returnColumn {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(QuoteIdent(col))
		}
	}

	return nil
}
