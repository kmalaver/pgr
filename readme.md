# interfaces

```go

  type DML interface {
    Select(...string) *SelectBuilder
    SelectSql(query string, args ...interface{}) *SelectBuilder
    InsertInto(string) *InsertBuilder
    InsertSql(query string, args ...interface{}) *InsertBuilder
    Update(string) *UpdateBuilder
    UpdateSql(query string, args ...interface{}) *UpdateBuilder
    DeleteFrom(string) *DeleteBuilder
    DeleteSql(query string, args ...interface{}) *DeleteBuilder
    Transaction(ctx context.Context, fn func(ctx context.Context) error) error
    With(name string, builder Builder) DML
  }

  type SelectBuilder interface {
    From(table interface{}) SelectBuilder
    Distinct() SelectBuilder
    Where(query interface{}, value ...interface{}) SelectBuilder
    Having(query interface{}, value ...interface{}) SelectBuilder
    GroupBy(colss ...string) SelectBuilder
    Limit(count uint64) SelectBuilder
    Offset(count uint64) SelectBuilder
    OrderBy(cols ...string) SelectBuilder
    OrderAsc(col string) SelectBuilder
    OrderDesc(col string) SelectBuilder
    Paginate(page, perPage int64) SelectBuilder
    OrderDir(col string, isAsc bool) SelectBuilder
    Join(table, on interface{}) SelectBuilder
    LeftJoin(table, on interface{}) SelectBuilder
    RightJoin(table, on interface{}) SelectBuilder
    FullJoin(table, on interface{}) SelectBuilder
    As(alias string) Builder
    Rows(ctx context.Context) (pgx.Rows, error)
    Load(ctx context.Context, dest interface{}) error
    LoadOne(ctx context.Context, dest interface{}) error
  }

  type ExecRunner interface {
    Exec(ctx context.Context) (int64, error)
    Load(ctx context.Context, dest interface{}) error
  }

  type InsertBuilder interface {
    Columns(columns ...string) InsertBuilder
    Values(values ...interface{}) InsertBuilder
    Pair(column string, value interface{}) InsertBuilder
    Returning(columns ...string) InsertBuilder
    Ignored() InsertBuilder
    Record(value interface{}) InsertBuilder
    Builder
    ExecRunner
  }

  type UpdateBuilder interface {
    Set(column string, value interface{}) UpdateBuilder
    Where(query interface{}, values ...interface{}) UpdateBuilder
    Returning(columns ...string) UpdateBuilder
    Builder
    ExecRunner
  }

  type DeleteBuilder interface {
    Where(query interface{}, values ...interface{}) DeleteBuilder
    Returning(columns ...string) DeleteBuilder
    Builder
    ExecRunner
  }
```
