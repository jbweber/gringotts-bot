package database_test

import (
	"testing"

	"github.com/jbweber/gringotts-bot/internal/database"
	"github.com/stretchr/testify/require"
)

func TestMigrator_Migrate(t *testing.T) {
	db, err := database.NewDB("file::memory:?cache=shared")
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	m := database.NewMigrator(db)
	err = m.Migrate()
	require.NoError(t, err)

	id, err := m.GetLatestMigrationID()
	require.NoError(t, err)
	require.Equal(t, 2, id)
}

func TestMigrator_GetLatestMigrationID(t *testing.T) {
	db, err := database.NewDB("file::memory:?cache=shared")
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	_, err = db.Exec(database.Migrations[1][0])
	require.NoError(t, err)

	m := database.NewMigrator(db)
	id, err := m.GetLatestMigrationID()
	require.NoError(t, err)
	require.Equal(t, -1, id)

	_, err = db.Exec("INSERT INTO migration (migration_id) values (1)")
	require.NoError(t, err)

	id, err = m.GetLatestMigrationID()
	require.NoError(t, err)
	require.Equal(t, 1, id)

	_, err = db.Exec("INSERT INTO migration (migration_id) values (2)")
	require.NoError(t, err)

	id, err = m.GetLatestMigrationID()
	require.NoError(t, err)
	require.Equal(t, 2, id)

	_, err = db.Exec("INSERT INTO migration (migration_id) values (3)")
	require.NoError(t, err)

	id, err = m.GetLatestMigrationID()
	require.NoError(t, err)
	require.Equal(t, 3, id)
}
