package pgr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelectBuilder(t *testing.T) {
	db := getDb()

	t.Run("select", func(t *testing.T) {
		buf := NewBuffer()
		b := db.Select("a", "b").
			From(
				Select("a").From("table").As("t1"),
			).
			LeftJoin("t2", "t1.a = t2.a").
			Distinct().
			Where(And(
				Or(Eq("c", 1), Like("c", "cc")),
				Expr("id in ?", []int64{1, 2, 3}),
			)).
			GroupBy("e").
			Having(Eq("f", 3)).
			OrderAsc("g").
			Limit(4).
			Offset(5)

		err := b.Build(buf)
		require.NoError(t, err)
		require.Equal(t, `SELECT DISTINCT a, b FROM ? `+
			`LEFT JOIN "t2" ON t1.a = t2.a `+
			`WHERE ((("c" = ?) OR ("c" LIKE 'cc')) AND (id in ?)) `+
			`GROUP BY e HAVING ("f" = ?) `+
			`ORDER BY g ASC LIMIT 4 OFFSET 5`,
			buf.String())
		require.Equal(t, 4, len(buf.Value()))
	})
}

func TestSelect(t *testing.T) {
	db := getDb()
	ctx := context.Background()
	_, err := db.InsertInto("users").
		Columns("name", "age").
		Values("a", 1).
		Values("b", 2).
		Exec(ctx)
	require.NoError(t, err)

	want := []User{
		{Name: "a", Age: 1},
		{Name: "b", Age: 2},
	}

	actual := []User{}

	db.Select("id", "name", "age").
		From("users").
		OrderAsc("id").
		Load(ctx, &actual)

	require.Equal(t, len(want), len(actual))
	for i, row := range actual {
		require.NotNil(t, row.Id)
		require.Equal(t, want[i].Name, row.Name)
		require.Equal(t, want[i].Age, row.Age)
	}
}

func TestSelectLoad(t *testing.T) {
	db := getDb()
	ctx := context.Background()

	var uIds []int64
	err := db.InsertInto("users").
		Columns("name", "age").
		Values("user a", 1).
		Values("user b", 2).
		Returning("id").
		Load(ctx, &uIds)
	require.NoError(t, err)

	var mIds []int64
	err = db.InsertInto("movies").
		Columns("name", "description").
		Values("movie a", "description a").
		Values("movie b", nil).
		Returning("id").
		Load(ctx, &mIds)
	require.NoError(t, err)

	_, err = db.InsertInto("user_movies").
		Columns("user_id", "movie_id").
		Values(uIds[0], mIds[0]).
		Values(uIds[0], mIds[1]).
		Values(uIds[1], mIds[0]).
		Exec(ctx)
	require.NoError(t, err)

	var users []User
	_, err = db.Select(
		"users.id",
		"users.name",
		"users.age",
		"json_agg(movies) AS movies",
	).
		From("users").
		LeftJoin(I("user_movies").As("um"), "um.user_id = users.id").
		LeftJoin("movies", "movies.id = um.movie_id").
		OrderAsc("users.id").
		GroupBy("users.id").
		Load(ctx, &users)

	expected := []User{
		{
			Id:   uIds[0],
			Name: "user a",
			Age:  1,
			Movies: []Movie{
				{
					Id:          mIds[0],
					Name:        "movie a",
					Description: strPtr("description a"),
				},
				{
					Id:          mIds[1],
					Name:        "movie b",
					Description: nil,
				},
			},
		},
		{
			Id:   uIds[1],
			Name: "user b",
			Age:  2,
			Movies: []Movie{
				{
					Id:          mIds[0],
					Name:        "movie a",
					Description: strPtr("description a"),
				},
			},
		},
	}

	require.NoError(t, err)
	require.Equal(t, expected, users)
}

func strPtr(s string) *string {
	return &s
}
