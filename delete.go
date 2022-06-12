package pgr

import (
	"context"
	"strconv"
)

type DeleteBuilder struct {
	db *Pgr
	raw

	table        string
	whereCond    []Builder
	limitCount   int64
	returnColumn []string
}

func (db *Pgr) DeleteFrom(table string) *DeleteBuilder {
	return &DeleteBuilder{
		table:      table,
		limitCount: -1,
		db:         db,
	}
}

func (db *Pgr) DeleteSql(query string, value ...interface{}) *DeleteBuilder {
	return &DeleteBuilder{
		raw: raw{
			Query: query,
			Value: value,
		},
		limitCount: -1,
		db:         db,
	}
}

func (b *DeleteBuilder) Where(query interface{}, value ...interface{}) *DeleteBuilder {
	switch query := query.(type) {
	case string:
		b.whereCond = append(b.whereCond, Expr(query, value...))
	case Builder:
		b.whereCond = append(b.whereCond, query)
	}
	return b
}

func (b *DeleteBuilder) Returning(columns ...string) *DeleteBuilder {
	b.returnColumn = columns
	return b
}

func (b *DeleteBuilder) Exec(ctx context.Context) (int64, error) {
	return b.db.exec(ctx, b)
}

func (b *DeleteBuilder) Load(ctx context.Context, dest interface{}) error {
	_, err := b.db.query(ctx, b, dest)
	return err
}

func (b *DeleteBuilder) Build(buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.Build(buf)
	}

	if b.table == "" {
		return ErrTableNotSpecified
	}

	buf.WriteString("DELETE FROM ")
	buf.WriteString(QuoteIdent(b.table))

	if len(b.whereCond) > 0 {
		buf.WriteString(" WHERE ")
		err := And(b.whereCond...).Build(buf)
		if err != nil {
			return err
		}
	}

	if b.limitCount >= 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.FormatInt(b.limitCount, 10))
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
