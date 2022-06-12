package pgr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInsertBuilder(t *testing.T) {
	db := getDb()

	t.Run("insert with columns and values", func(t *testing.T) {
		buf := NewBuffer()
		err := db.InsertInto("users").
			Columns("name", "age").
			Values("a", 1).
			Values("b", 2).
			Values("c", 3).
			Returning("id", "name", "age").
			Build(buf)
		require.NoError(t, err)
		require.Equal(t, `INSERT INTO "users" ("name","age") VALUES (?,?), (?,?), (?,?) RETURNING "id","name","age"`, buf.String())
		require.Equal(t, []interface{}{"a", 1, "b", 2, "c", 3}, buf.Value())
	})

	t.Run("insert with pair", func(t *testing.T) {
		buf := NewBuffer()
		err := db.InsertInto("users").
			Pair("name", "a").
			Pair("age", 1).
			Build(buf)
		require.NoError(t, err)
		require.Equal(t, `INSERT INTO "users" ("name","age") VALUES (?,?)`, buf.String())
		require.Equal(t, []interface{}{"a", 1}, buf.Value())
	})

	t.Run("insert with record", func(t *testing.T) {
		buf := NewBuffer()
		user := User{
			Name: "a",
			Age:  1,
		}
		err := db.InsertInto("users").Columns("name", "age").Record(&user).Build(buf)
		require.NoError(t, err)
		require.Equal(t, `INSERT INTO "users" ("name","age") VALUES (?,?)`, buf.String())
		require.Equal(t, []interface{}{"a", 1}, buf.Value())
	})
}

func TestInsert(t *testing.T) {
	db := getDb()
	ctx := context.Background()

	var ids []int64
	err := db.InsertInto("users").
		Columns("name", "age").
		Values("a", 1).
		Values("b", 2).
		Returning("id").
		Load(ctx, &ids)

	require.NoError(t, err)
	require.Equal(t, 2, len(ids))

	var count Count
	db.Select("COUNT(*)").From("users").LoadOne(ctx, &count)
	require.Equal(t, 2, count.Count)
}
