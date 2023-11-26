package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Gringotts struct {
	db *sql.DB
}

func NewGringotts(db *sql.DB) *Gringotts {
	return &Gringotts{db: db}
}

type Item struct {
	ID    string
	Name  string
	Count int
}

func (g *Gringotts) FindItem(ctx context.Context, searchString string) ([]*Item, error) {
	query := fmt.Sprintf(`
		SELECT i.id, i.name, SUM(ic.item_count) as item_total FROM item i
		LEFT JOIN item_count ic
		ON i.id = ic.item_id
		WHERE i.name LIKE '%%%s%%'
		GROUP BY i.name
		`, searchString,
	)

	r, err := g.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer func() { _ = r.Close() }()

	var items []*Item
	for r.Next() {
		i := &Item{}
		if err := r.Scan(&i.ID, &i.Name, &i.Count); err != nil {
			return nil, err
		}

		items = append(items, i)
	}

	return items, nil
}

func (g *Gringotts) GetItemCount(ctx context.Context, owner string, itemID int) (int, error) {
	stmt, err := g.db.PrepareContext(ctx, `SELECT item_count FROM item_count WHERE owner = ? AND item_id = ?`)
	if err != nil {
		return -1, err
	}

	defer func() { _ = stmt.Close() }() // TODO better

	r := stmt.QueryRowContext(ctx, owner, itemID)

	var count int
	err = r.Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// TODO handle this with a sentinal?
		}
		return -1, err
	}

	return count, nil
}

func (g *Gringotts) GetItemName(ctx context.Context, id string) (string, error) {
	stmt, err := g.db.PrepareContext(ctx, `SELECT name FROM item WHERE id = ?`)
	if err != nil {
		return "", err
	}

	defer func() { _ = stmt.Close() }() // TODO better

	r := stmt.QueryRowContext(ctx, id)

	var name string
	err = r.Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// TODO handle this with a sentinal?
		}
		return "", err
	}

	return name, nil
}

func (g *Gringotts) UpdateItemCounts(ctx context.Context, owner string, itemCounts map[string]int) error {
	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt1, err := tx.PrepareContext(ctx, `DELETE FROM item_count where owner = ?`)
	if err != nil {
		return err
	}

	_, err = stmt1.ExecContext(ctx, owner)
	if err != nil {
		return err
	}

	defer func() { _ = stmt1.Close() }() // TODO better

	stmt2, err := tx.PrepareContext(ctx, `INSERT INTO item_count (owner, item_id, item_count) VALUES (?,?,?)`)
	if err != nil {
		return err
	}

	defer func() { _ = stmt2.Close() }() // TODO better

	for k, v := range itemCounts {
		_, err := stmt2.ExecContext(ctx, owner, k, v)
		if err != nil {
			_ = tx.Rollback() // TODO multierr
			return err
		}
	}

	err = tx.Commit()

	return err
}

func (g *Gringotts) UpdateItems(ctx context.Context, items map[string]string) error {
	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO item (id, name) values(?, ?) ON CONFLICT(id) DO UPDATE SET name = excluded.name WHERE id = excluded.id`)
	if err != nil {
		return err
	}

	defer func() { _ = stmt.Close() }() // TODO better

	for k, v := range items {
		_, err := stmt.ExecContext(ctx, k, v)
		if err != nil {
			_ = tx.Rollback() // TODO multierr
			return err
		}
	}

	err = tx.Commit()

	return err
}
