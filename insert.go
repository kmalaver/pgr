package pgr

import (
	"context"
	"reflect"
	"strings"
)

type InsertBuilder struct {
	db *Pgr
	raw

	table        string
	columns      []string
	returnColumn []string
	values       [][]interface{}
	ignored      bool
}

func (db *Pgr) InsertInto(table string) *InsertBuilder {
	return &InsertBuilder{
		table: table,
		db:    db,
	}
}

func (db *Pgr) InsertSql(query string, value ...interface{}) *InsertBuilder {
	return &InsertBuilder{
		raw: raw{
			Query: query,
			Value: value,
		},
		db: db,
	}
}

func (b *InsertBuilder) Columns(columns ...string) *InsertBuilder {
	b.columns = columns
	return b
}

func (b *InsertBuilder) Values(values ...interface{}) *InsertBuilder {
	b.values = append(b.values, values)
	return b
}

func (b *InsertBuilder) Pair(column string, value interface{}) *InsertBuilder {
	b.columns = append(b.columns, column)
	switch len(b.values) {
	case 0:
		b.Values(value)
	case 1:
		b.values[0] = append(b.values[0], value)
	default:
		panic("pair only allows one record to insert")
	}
	return b
}

func (b *InsertBuilder) Returning(columns ...string) *InsertBuilder {
	b.returnColumn = columns
	return b
}

// Record is a helper function to insert a single record. using a struct as the value.
//
// If no Columns are specified, the columns will be set by the
// struct fields excluding non exported fields.
func (b *InsertBuilder) Record(record interface{}) *InsertBuilder {
	v := reflect.Indirect(reflect.ValueOf(record))

	if v.Kind() == reflect.Struct {
		s := newTagStore()

		// if no columns are specified, use the struct fields
		if len(b.columns) == 0 {
			fields := s.get(v.Type())
			for i, field := range fields {
				if field == "id" {
					// skip id field
					if idField := v.Field(i); idField.IsZero() {
						continue
					}
				}
				if field != "" {
					b.columns = append(b.columns, field)
				}
			}
		}

		// TODO: check this code
		found := make([]interface{}, len(b.columns)+1)
		s.findValueByName(v, append(b.columns, "id"), found, false)

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
	b.ignored = true
	return b
}

func (b *InsertBuilder) Exec(ctx context.Context) (int64, error) {
	return b.db.exec(ctx, b)
}

func (b *InsertBuilder) Load(ctx context.Context, dest interface{}) error {
	_, err := b.db.query(ctx, b, dest)
	return err
}

func (b *InsertBuilder) Build(buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.Build(buf)
	}

	if b.table == "" {
		return ErrTableNotSpecified
	}

	if len(b.columns) == 0 {
		return ErrColumnNotSpecified
	}

	if b.ignored {
		buf.WriteString("INSERT IGNORE INTO ")
	} else {
		buf.WriteString("INSERT INTO ")
	}

	buf.WriteString(QuoteIdent(b.table))

	var placeholderBuf strings.Builder
	placeholderBuf.WriteString("(")
	buf.WriteString(" (")
	for i, col := range b.columns {
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

	for i, tuple := range b.values {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(placeholderStr)

		buf.WriteValue(tuple...)
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
