package queryx

import (
	"context"
	"strconv"
)

type DeleteBuilder struct {
	runner runner
	raw
	Table      string
	WhereCond  []Builder
	LimitCount int64
}

func (db queryx) DeleteFrom(table string) *DeleteBuilder {
	return &DeleteBuilder{
		Table:      table,
		LimitCount: -1,
		runner:     db.conn,
	}
}

func (db queryx) DeleteSql(query string, value ...interface{}) *DeleteBuilder {
	return &DeleteBuilder{
		raw: raw{
			Query: query,
			Value: value,
		},
		LimitCount: -1,
		runner:     db.conn,
	}
}

func (b *DeleteBuilder) Where(query interface{}, value ...interface{}) *DeleteBuilder {
	switch query := query.(type) {
	case string:
		b.WhereCond = append(b.WhereCond, Expr(query, value...))
	case Builder:
		b.WhereCond = append(b.WhereCond, query)
	}
	return b
}

func (b *DeleteBuilder) Limit(n uint64) *DeleteBuilder {
	b.LimitCount = int64(n)
	return b
}

func (b *DeleteBuilder) Exec(ctx context.Context) (int64, error) {
	res, err := exec(ctx, b.runner, b)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (b *DeleteBuilder) Build(buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.Build(buf)
	}

	if b.Table == "" {
		return ErrTableNotSpecified
	}

	buf.WriteString("DELETE FROM ")
	buf.WriteString(QuoteIdent(b.Table))

	if len(b.WhereCond) > 0 {
		buf.WriteString(" WHERE ")
		err := And(b.WhereCond...).Build(buf)
		if err != nil {
			return err
		}
	}
	if b.LimitCount >= 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.FormatInt(b.LimitCount, 10))
	}
	return nil
}
