package database_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/jbweber/gringotts-bot/internal/database"
	"github.com/stretchr/testify/require"
)

func getGringotts(t *testing.T) (*database.Gringotts, *sql.DB) {
	db, err := database.NewDB(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name()))
	require.NoError(t, err)

	migrator := database.NewMigrator(db)
	err = migrator.Migrate()
	require.NoError(t, err)

	return database.NewGringotts(db), db
}

var (
	items1 = map[string]string{
		"1": "item 1",
		"2": "item 2",
		"3": "item 3",
		"4": "item 4",
	}
	items2 = map[string]string{
		"1": "item 1",
		"2": "item 2",
		"3": "item 33",
		"4": "item 44",
		"5": "item 5",
	}
	itemCounts1 = map[string]int{
		"1": 1,
		"2": 2,
		"3": 3,
		"4": 4,
		"5": 5,
	}
	itemCounts2 = map[string]int{
		"1": 5,
		"2": 4,
		"3": 2,
		"4": 1,
		"5": 3,
	}
)

func TestGringotts_GetItemName(t *testing.T) {
	g, db := getGringotts(t)

	defer func() { _ = db.Close() }()

	err := g.UpdateItems(context.Background(), items1)
	require.NoError(t, err)

	name, err := g.GetItemName(context.Background(), "1")
	require.NoError(t, err)
	require.Equal(t, "item 1", name)
}

func TestGringotts_UpdateItemCounts(t *testing.T) {
	g, db := getGringotts(t)

	defer func() { _ = db.Close() }()

	testOwner := "testChar"

	err := g.UpdateItemCounts(context.Background(), testOwner, itemCounts1)
	require.NoError(t, err)

	r, err := db.Query(fmt.Sprintf("SELECT COUNT(id) FROM item_count WHERE owner='%s'", testOwner))
	require.NoError(t, err)

	for r.Next() {
		var count int
		err = r.Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 5, count)
	}
	err = r.Close()
	require.NoError(t, err)

	one, err := g.GetItemCount(context.Background(), testOwner, 1)
	require.NoError(t, err)
	require.Equal(t, itemCounts1["1"], one)

	two, err := g.GetItemCount(context.Background(), testOwner, 2)
	require.NoError(t, err)
	require.Equal(t, itemCounts1["2"], two)

	three, err := g.GetItemCount(context.Background(), testOwner, 3)
	require.NoError(t, err)
	require.Equal(t, itemCounts1["3"], three)

	four, err := g.GetItemCount(context.Background(), testOwner, 4)
	require.NoError(t, err)
	require.Equal(t, itemCounts1["4"], four)

	five, err := g.GetItemCount(context.Background(), testOwner, 5)
	require.NoError(t, err)
	require.Equal(t, itemCounts1["5"], five)

	err = g.UpdateItemCounts(context.Background(), testOwner, itemCounts2)
	require.NoError(t, err)

	r, err = db.Query(fmt.Sprintf("SELECT COUNT(id) FROM item_count WHERE owner='%s'", testOwner))
	require.NoError(t, err)

	for r.Next() {
		var count int
		err = r.Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 5, count)
	}

	err = r.Close()
	require.NoError(t, err)

	one, err = g.GetItemCount(context.Background(), testOwner, 1)
	require.NoError(t, err)
	require.Equal(t, itemCounts2["1"], one)

	two, err = g.GetItemCount(context.Background(), testOwner, 2)
	require.NoError(t, err)
	require.Equal(t, itemCounts2["2"], two)

	three, err = g.GetItemCount(context.Background(), testOwner, 3)
	require.NoError(t, err)
	require.Equal(t, itemCounts2["3"], three)

	four, err = g.GetItemCount(context.Background(), testOwner, 4)
	require.NoError(t, err)
	require.Equal(t, itemCounts2["4"], four)

	five, err = g.GetItemCount(context.Background(), testOwner, 5)
	require.NoError(t, err)
	require.Equal(t, itemCounts2["5"], five)
}

func TestGringotts_UpdateItems(t *testing.T) {
	g, db := getGringotts(t)

	defer func() { _ = db.Close() }()

	err := g.UpdateItems(context.Background(), items1)
	require.NoError(t, err)

	r, err := db.Query("SELECT COUNT(id) FROM item")
	require.NoError(t, err)
	defer func() { _ = r.Close() }()

	for r.Next() {
		var count int
		err = r.Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 4, count)
	}

	err = g.UpdateItems(context.Background(), items2)
	require.NoError(t, err)

	r, err = db.Query("SELECT COUNT(id) FROM item")
	require.NoError(t, err)
	defer func() { _ = r.Close() }()

	for r.Next() {
		var count int
		err = r.Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 5, count)
	}
	one, err := g.GetItemName(context.Background(), "1")
	require.NoError(t, err)
	require.Equal(t, items2["1"], one)

	two, err := g.GetItemName(context.Background(), "2")
	require.NoError(t, err)
	require.Equal(t, items2["2"], two)

	three, err := g.GetItemName(context.Background(), "3")
	require.NoError(t, err)
	require.Equal(t, items2["3"], three)

	four, err := g.GetItemName(context.Background(), "4")
	require.NoError(t, err)
	require.Equal(t, items2["4"], four)

	five, err := g.GetItemName(context.Background(), "5")
	require.NoError(t, err)
	require.Equal(t, items2["5"], five)
}
