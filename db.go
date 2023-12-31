package main

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

type Database struct {
	sql *sqlx.DB
}

func initDatabase(conn string) (*Database, error) {
	db, err := sqlx.Open("postgres", conn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Database{sql: db}, nil
}

func (db *Database) updateBalance(ctx context.Context, bl *balanceLimit) (*balanceLimit, error) {
	updatedAt := time.Now()

	row := db.sql.QueryRowxContext(
		ctx,
		`UPDATE balance SET balance = $1, status = $2, day_limit = $3, updated_at = $4, today_spent = $5, today_added = $6 WHERE name='VladyaPolya' RETURNING *`,
		bl.Balance,
		bl.Status,
		bl.DayLimit,
		updatedAt,
		bl.TodaySpent,
		bl.TodayAdded,
	)

	var updated balanceLimit
	if err := row.StructScan(&updated); err != nil {
		return nil, fmt.Errorf("scan sql result: %w", err)
	}

	return &updated, nil
}

func (db *Database) getBalance(ctx context.Context) (*balanceLimit, error) {
	var bl balanceLimit

	row := db.sql.QueryRowxContext(ctx, `SELECT * FROM balance WHERE name='VladyaPolya'`)
	if err := row.StructScan(&bl); err != nil {
		return nil, fmt.Errorf("scan sql result: %w", err)
	}

	return &bl, nil
}
