package pgr

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v4"
)

type SelectBuilder struct {
	db *Pgr
	raw

	distinct bool

	columns    []interface{}
	table      interface{}
	joinTables []Builder

	whereCond  []Builder
	group      []Builder
	havingCond []Builder
	order      []Builder

	limitCount  int64
	offsetCount int64
}

func prepareSelect(a []string) []interface{} {
	b := make([]interface{}, len(a))
	for i := range a {
		b[i] = a[i]
	}
	return b
}

// Select creates a SelectBuilder.
func Select(cols ...interface{}) *SelectBuilder {
	return &SelectBuilder{
		columns:     cols,
		limitCount:  -1,
		offsetCount: -1,
	}
}

// Select creates a SelectBuilder.
func (db *Pgr) Select(cols ...string) *SelectBuilder {
	b := Select(prepareSelect(cols)...)
	b.db = db
	return b
}

// SelectSql creates a SelectBuilder with raw SQL.
func (db *Pgr) SelectSql(query string, value ...interface{}) *SelectBuilder {
	return &SelectBuilder{
		raw: raw{
			Query: query,
			Value: value,
		},
		limitCount:  -1,
		offsetCount: -1,
		db:          db,
	}
}

// From specifies table to select from.
// table can be Builder or string.
func (b *SelectBuilder) From(table interface{}) *SelectBuilder {
	b.table = table
	return b
}

// Distinct adds DISTINCT clause.
func (b *SelectBuilder) Distinct() *SelectBuilder {
	b.distinct = true
	return b
}

// Where adds a where condition.
// query can be Builder or string. value is used only if query type is string.
func (b *SelectBuilder) Where(query interface{}, value ...interface{}) *SelectBuilder {
	switch query := query.(type) {
	case string:
		b.whereCond = append(b.whereCond, Expr(query, value...))
	case Builder:
		b.whereCond = append(b.whereCond, query)
	}
	return b
}

// Having adds a having condition.
// query can be Builder or string. value is used only if query type is string.
func (b *SelectBuilder) Having(query interface{}, value ...interface{}) *SelectBuilder {
	switch query := query.(type) {
	case string:
		b.havingCond = append(b.havingCond, Expr(query, value...))
	case Builder:
		b.havingCond = append(b.havingCond, query)
	}
	return b
}

// GroupBy specifies columns for grouping.
func (b *SelectBuilder) GroupBy(col ...string) *SelectBuilder {
	for _, group := range col {
		b.group = append(b.group, Expr(group))
	}
	return b
}

// OrderAsc adds a column to the ORDER BY clause.
func (b *SelectBuilder) OrderAsc(col string) *SelectBuilder {
	b.order = append(b.order, order(col, asc))
	return b
}

// OrderDesc adds a column to the ORDER BY clause.
func (b *SelectBuilder) OrderDesc(col string) *SelectBuilder {
	b.order = append(b.order, order(col, desc))
	return b
}

// OrderBy specifies columns for ordering.
func (b *SelectBuilder) OrderBy(col string) *SelectBuilder {
	b.order = append(b.order, Expr(col))
	return b
}

func (b *SelectBuilder) Limit(n uint64) *SelectBuilder {
	b.limitCount = int64(n)
	return b
}

func (b *SelectBuilder) Offset(n uint64) *SelectBuilder {
	b.offsetCount = int64(n)
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

// Join add inner join.
// on can be Builder or string.
func (b *SelectBuilder) Join(table, on interface{}) *SelectBuilder {
	b.joinTables = append(b.joinTables, join(inner, table, on))
	return b
}

// LeftJoin add left join.
// on can be Builder or string.
func (b *SelectBuilder) LeftJoin(table, on interface{}) *SelectBuilder {
	b.joinTables = append(b.joinTables, join(left, table, on))
	return b
}

// RightJoin add right join.
// on can be Builder or string.
func (b *SelectBuilder) RightJoin(table, on interface{}) *SelectBuilder {
	b.joinTables = append(b.joinTables, join(right, table, on))
	return b
}

// FullJoin add full join.
// on can be Builder or string.
func (b *SelectBuilder) FullJoin(table, on interface{}) *SelectBuilder {
	b.joinTables = append(b.joinTables, join(full, table, on))
	return b
}

// As creates alias for select statement.
func (b *SelectBuilder) As(alias string) Builder {
	return as(b, alias)
}

// Rows executes the query and returns a Rows object.
func (b *SelectBuilder) Rows(ctx context.Context) (pgx.Rows, error) {
	_, rows, err := b.db.queryRows(ctx, b)
	return rows, err
}

// LoadOne executes the query and loads one record into given struct.
func (b *SelectBuilder) LoadOne(ctx context.Context, dest interface{}) error {
	count, err := b.db.query(ctx, b, dest)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

// Load executes the query and loads all records into given struct.
func (b *SelectBuilder) Load(ctx context.Context, dest interface{}) (int, error) {
	return b.db.query(ctx, b, dest)
}

func (b *SelectBuilder) Build(buf Buffer) error {
	if b.raw.Query != "" {
		return b.raw.Build(buf)
	}

	if len(b.columns) == 0 {
		return ErrColumnNotSpecified
	}

	buf.WriteString("SELECT ")

	if b.distinct {
		buf.WriteString("DISTINCT ")
	}

	for i, col := range b.columns {
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

	if b.table != nil {
		buf.WriteString(" FROM ")
		switch table := b.table.(type) {
		case string:
			buf.WriteString(table)
		default:
			buf.WriteString(placeholder)
			buf.WriteValue(table)
		}

		if len(b.joinTables) > 0 {
			for _, join := range b.joinTables {
				err := join.Build(buf)
				if err != nil {
					return err
				}
			}
		}
	}

	if len(b.whereCond) > 0 {
		buf.WriteString(" WHERE ")
		err := And(b.whereCond...).Build(buf)
		if err != nil {
			return err
		}
	}

	if len(b.group) > 0 {
		buf.WriteString(" GROUP BY ")
		for i, group := range b.group {
			if i > 0 {
				buf.WriteString(", ")
			}
			err := group.Build(buf)
			if err != nil {
				return err
			}
		}
	}

	if len(b.havingCond) > 0 {
		buf.WriteString(" HAVING ")
		err := And(b.havingCond...).Build(buf)
		if err != nil {
			return err
		}
	}

	if len(b.order) > 0 {
		buf.WriteString(" ORDER BY ")
		for i, order := range b.order {
			if i > 0 {
				buf.WriteString(", ")
			}
			err := order.Build(buf)
			if err != nil {
				return err
			}
		}
	}

	if b.limitCount >= 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.FormatInt(b.limitCount, 10))
	}

	if b.offsetCount >= 0 {
		buf.WriteString(" OFFSET ")
		buf.WriteString(strconv.FormatInt(b.offsetCount, 10))
	}

	return nil
}
