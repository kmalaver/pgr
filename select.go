package queryx

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v4"
)

type SelectBuilder struct {
	runner runner
	raw

	IsDistinct bool

	Column    []interface{}
	Table     interface{}
	JoinTable []Builder

	WhereCond  []Builder
	Group      []Builder
	HavingCond []Builder
	Order      []Builder
	Suffixes   []Builder

	LimitCount  int64
	OffsetCount int64
}

func prepareSelect(a []string) []interface{} {
	b := make([]interface{}, len(a))
	for i := range a {
		b[i] = a[i]
	}
	return b
}

func Select(cols ...string) *SelectBuilder {
	return &SelectBuilder{
		Column:      prepareSelect(cols),
		LimitCount:  -1,
		OffsetCount: -1,
	}
}

func (db *queryx) Select(cols ...string) *SelectBuilder {
	b := Select(cols...)
	b.runner = db.conn
	return b
}

func (db *queryx) SelectSql(query string, value ...interface{}) *SelectBuilder {
	return &SelectBuilder{
		raw: raw{
			Query: query,
			Value: value,
		},
		LimitCount:  -1,
		OffsetCount: -1,
		runner:      db.conn,
	}
}

// From specifies table to select from.
// table can be Builder like SelectBuilder, or string.
func (b *SelectBuilder) From(table interface{}) *SelectBuilder {
	b.Table = table
	return b
}

func (b *SelectBuilder) Distinct() *SelectBuilder {
	b.IsDistinct = true
	return b
}

// Where adds a where condition.
// query can be Builder or string. value is used only if query type is string.
func (b *SelectBuilder) Where(query interface{}, value ...interface{}) *SelectBuilder {
	switch query := query.(type) {
	case string:
		b.WhereCond = append(b.WhereCond, Expr(query, value...))
	case Builder:
		b.WhereCond = append(b.WhereCond, query)
	}
	return b
}

// Having adds a having condition.
// query can be Builder or string. value is used only if query type is string.
func (b *SelectBuilder) Having(query interface{}, value ...interface{}) *SelectBuilder {
	switch query := query.(type) {
	case string:
		b.HavingCond = append(b.HavingCond, Expr(query, value...))
	case Builder:
		b.HavingCond = append(b.HavingCond, query)
	}
	return b
}

// GroupBy specifies columns for grouping.
func (b *SelectBuilder) GroupBy(col ...string) *SelectBuilder {
	for _, group := range col {
		b.Group = append(b.Group, Expr(group))
	}
	return b
}

func (b *SelectBuilder) OrderAsc(col string) *SelectBuilder {
	b.Order = append(b.Order, order(col, asc))
	return b
}

func (b *SelectBuilder) OrderDesc(col string) *SelectBuilder {
	b.Order = append(b.Order, order(col, desc))
	return b
}

// OrderBy specifies columns for ordering.
func (b *SelectBuilder) OrderBy(col string) *SelectBuilder {
	b.Order = append(b.Order, Expr(col))
	return b
}

func (b *SelectBuilder) Limit(n uint64) *SelectBuilder {
	b.LimitCount = int64(n)
	return b
}

func (b *SelectBuilder) Offset(n uint64) *SelectBuilder {
	b.OffsetCount = int64(n)
	return b
}

// Suffix adds an expression to the end of the query. This is useful to add dialect-specific clauses like FOR UPDATE
func (b *SelectBuilder) Suffix(suffix string, value ...interface{}) *SelectBuilder {
	b.Suffixes = append(b.Suffixes, Expr(suffix, value...))
	return b
}

// Paginate fetches a page in a naive way for a small set of data.
func (b *SelectBuilder) Paginate(page, perPage uint64) *SelectBuilder {
	b.Limit(perPage)
	b.Offset((page - 1) * perPage)
	return b
}

// OrderDir is a helper for OrderAsc and OrderDesc.
func (b *SelectBuilder) OrderDir(col string, isAsc bool) *SelectBuilder {
	if isAsc {
		b.OrderAsc(col)
	} else {
		b.OrderDesc(col)
	}
	return b
}

// Join add inner-join.
// on can be Builder or string.
func (b *SelectBuilder) Join(table, on interface{}) *SelectBuilder {
	b.JoinTable = append(b.JoinTable, join(inner, table, on))
	return b
}

// LeftJoin add left-join.
// on can be Builder or string.
func (b *SelectBuilder) LeftJoin(table, on interface{}) *SelectBuilder {
	b.JoinTable = append(b.JoinTable, join(left, table, on))
	return b
}

// RightJoin add right-join.
// on can be Builder or string.
func (b *SelectBuilder) RightJoin(table, on interface{}) *SelectBuilder {
	b.JoinTable = append(b.JoinTable, join(right, table, on))
	return b
}

// FullJoin add full-join.
// on can be Builder or string.
func (b *SelectBuilder) FullJoin(table, on interface{}) *SelectBuilder {
	b.JoinTable = append(b.JoinTable, join(full, table, on))
	return b
}

// As creates alias for select statement.
func (b *SelectBuilder) As(alias string) Builder {
	return as(b, alias)
}

func (b *SelectBuilder) Rows(ctx context.Context) (pgx.Rows, error) {
	return queryRows(ctx, b.runner, b)
}

func (b *SelectBuilder) LoadOne(ctx context.Context, value interface{}) error {
	count, err := query(ctx, b.runner, b, value)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (b *SelectBuilder) Load(ctx context.Context, value interface{}) (int, error) {
	return query(ctx, b.runner, b, value)
}

func (b *SelectBuilder) Build(buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.Build(buf)
	}

	if len(b.Column) == 0 {
		return ErrColumnNotSpecified
	}

	buf.WriteString("SELECT ")

	if b.IsDistinct {
		buf.WriteString("DISTINCT ")
	}

	for i, col := range b.Column {
		if i > 0 {
			buf.WriteString(", ")
		}
		switch col := col.(type) {
		case string:
			buf.WriteString(col)
		default:
			buf.WriteString(placeholder)
			buf.WriteValue(col)
		}
	}

	if b.Table != nil {
		buf.WriteString(" FROM ")
		switch table := b.Table.(type) {
		case string:
			buf.WriteString(table)
		default:
			buf.WriteString(placeholder)
			buf.WriteValue(table)
		}

		if len(b.JoinTable) > 0 {
			for _, join := range b.JoinTable {
				err := join.Build(buf)
				if err != nil {
					return err
				}
			}
		}
	}

	if len(b.WhereCond) > 0 {
		buf.WriteString(" WHERE ")
		err := And(b.WhereCond...).Build(buf)
		if err != nil {
			return err
		}
	}

	if len(b.Group) > 0 {
		buf.WriteString(" GROUP BY ")
		for i, group := range b.Group {
			if i > 0 {
				buf.WriteString(", ")
			}
			err := group.Build(buf)
			if err != nil {
				return err
			}
		}
	}

	if len(b.HavingCond) > 0 {
		buf.WriteString(" HAVING ")
		err := And(b.HavingCond...).Build(buf)
		if err != nil {
			return err
		}
	}

	if len(b.Order) > 0 {
		buf.WriteString(" ORDER BY ")
		for i, order := range b.Order {
			if i > 0 {
				buf.WriteString(", ")
			}
			err := order.Build(buf)
			if err != nil {
				return err
			}
		}
	}

	if b.LimitCount >= 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.FormatInt(b.LimitCount, 10))
	}

	if b.OffsetCount >= 0 {
		buf.WriteString(" OFFSET ")
		buf.WriteString(strconv.FormatInt(b.OffsetCount, 10))
	}

	if len(b.Suffixes) > 0 {
		for _, suffix := range b.Suffixes {
			buf.WriteString(" ")
			err := suffix.Build(buf)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
