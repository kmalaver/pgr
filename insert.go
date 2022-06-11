package queryx

import (
	"context"
	"reflect"
	"strings"
)

type InsertBuilder struct {
	runner runner
	raw
	Table        string
	Column       []string
	Value        [][]interface{}
	ReturnColumn []string
	Ignored      bool
}

func (db queryx) InsertInto(table string) *InsertBuilder {
	return &InsertBuilder{
		Table:  table,
		runner: db.conn,
	}
}

func (db queryx) InsertSql(query string, value ...interface{}) *InsertBuilder {
	return &InsertBuilder{
		raw: raw{
			Query: query,
			Value: value,
		},
		runner: db.conn,
	}
}

func (b *InsertBuilder) Columns(columns ...string) *InsertBuilder {
	b.Column = columns
	return b
}

func (b *InsertBuilder) Values(values ...interface{}) *InsertBuilder {
	b.Value = append(b.Value, values)
	return b
}

func (b *InsertBuilder) Pair(column string, value interface{}) *InsertBuilder {
	b.Column = append(b.Column, column)
	switch len(b.Value) {
	case 0:
		b.Values(value)
	case 1:
		b.Value[0] = append(b.Value[0], value)
	default:
		panic("pair only allows one record to insert")
	}
	return b
}

func (b *InsertBuilder) Returning(columns ...string) *InsertBuilder {
	b.ReturnColumn = columns
	return b
}

func (b *InsertBuilder) Record(record interface{}) *InsertBuilder {
	v := reflect.Indirect(reflect.ValueOf(record))

	if v.Kind() == reflect.Struct {
		found := make([]interface{}, len(b.Column)+1)
		s := newTagStore()
		s.findValueByName(v, append(b.Column, "id"), found, false)

		value := found[:len(found)-1]
		for i, v := range value {
			if v != nil {
				value[i] = v.(reflect.Value).Interface()
			}
		}
		b.Values(value...)
	}
	return b
}

func (b *InsertBuilder) Ignore() *InsertBuilder {
	b.Ignored = true
	return b
}

func (b *InsertBuilder) Exec(ctx context.Context) (int64, error) {
	result, err := exec(ctx, b.runner, b)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (b *InsertBuilder) Load(ctx context.Context, value interface{}) error {
	_, err := query(ctx, b.runner, b, value)
	return err
}

func (b *InsertBuilder) Build(buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.Build(buf)
	}

	if b.Table == "" {
		return ErrTableNotSpecified
	}

	if len(b.Column) == 0 {
		return ErrColumnNotSpecified
	}

	if b.Ignored {
		buf.WriteString("INSERT IGNORE INTO ")
	} else {
		buf.WriteString("INSERT INTO ")
	}

	buf.WriteString(QuoteIdent(b.Table))

	var placeholderBuf strings.Builder
	placeholderBuf.WriteString("(")
	buf.WriteString(" (")
	for i, col := range b.Column {
		if i > 0 {
			buf.WriteString(",")
			placeholderBuf.WriteString(",")
		}
		buf.WriteString(QuoteIdent(col))
		placeholderBuf.WriteString(placeholder)
	}
	buf.WriteString(")")

	buf.WriteString(" VALUES ")
	placeholderBuf.WriteString(")")
	placeholderStr := placeholderBuf.String()

	for i, tuple := range b.Value {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(placeholderStr)

		buf.WriteValue(tuple...)
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

	return nil
}
