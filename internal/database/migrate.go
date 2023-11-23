package database

import (
	"database/sql"
	"sort"
)

var Migrations = map[int][]string{
	1: {
		`
		CREATE TABLE IF NOT EXISTS migration (
    		id INTEGER PRIMARY KEY NOT NULL,
    		updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
    		migration_id INT NOT NULL UNIQUE
		)
		`,
		`
		INSERT INTO migration (migration_id) values(1)
		`,
	},
	2: {
		`
		CREATE TABLE IF NOT EXISTS item_count (
		    id INTEGER PRIMARY KEY NOT NULL,
		    owner VARCHAR(64) NOT NULL,
		    item_id VARCHAR(64) NOT NULL COLLATE NOCASE,
		    item_count INTEGER NOT NULL DEFAULT 0,
		    FOREIGN KEY(item_id) REFERENCES item(id),
		    UNIQUE(owner, item_id)
		)
		`,
		`
		CREATE TABLE IF NOT EXISTS item (
		    id VARCHAR(64) PRIMARY KEY NOT NULL,
		    name VARCHAR(255) UNIQUE COLLATE NOCASE
		)
		`,
		`
		INSERT INTO migration (migration_id) values(2)
		`,
	},
}

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) GetLatestMigrationID() (int, error) {
	q := `SELECT migration_id FROM migration ORDER BY migration_id DESC LIMIT 1`
	r := m.db.QueryRow(q)
	var id int
	switch err := r.Scan(&id); err {
	case sql.ErrNoRows:
		return -1, nil
	case nil:
		return id, nil
	default:
		return -1, err
	}
}

func (m *Migrator) Migrate() error {
	latest, err := m.GetLatestMigrationID()
	if err != nil {
		if err.Error() != "no such table: migration" {
			return err
		}
	}

	var ids []int
	for k := range Migrations {
		if k > latest {
			ids = append(ids, k)
		}
	}

	sort.Ints(ids)

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	for _, id := range ids {
		for _, q := range Migrations[id] {
			if _, err := tx.Exec(q); err != nil {
				tx.Rollback() // TODO multierr
				return err
			}
		}
	}

	err = tx.Commit()

	return err
}
