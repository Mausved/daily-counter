package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"math"
	"strings"
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

func (db *Database) updateOnlyBalance(
	ctx context.Context,
	bl *balanceLimit,
) (*balanceLimit, error) {
	return db.updateBalance(ctx, bl, 0, "")
}

func (db *Database) updateBalance(
	ctx context.Context,
	bl *balanceLimit,
	delta float64,
	tag string,
) (*balanceLimit, error) {
	updatedAt := time.Now()

	if delta != 0 {
		uuid := uuid.NewString()

		_, err := db.sql.Queryx(
			`INSERT INTO transactions (uuid, balance_id, delta, date, tag) VALUES ($1, $2, $3, $4, $5)`,
			uuid,
			bl.Id,
			delta,
			time.Now().Format(time.DateOnly),
			tag,
		)

		if err != nil {
			return nil, fmt.Errorf("queryx: %w", err)
		}
	}

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

func (db *Database) getTransactionsForMonth(_ context.Context, balanceID int) (map[string]float64, error) {
	yy, mm, _ := time.Now().Date()

	lastDay := monthLastDay(time.Now())

	compareStringBefore := fmt.Sprintf(`%d-%d-0%d`, yy, mm, 1)
	compareStringUntil := fmt.Sprintf(`%d-%d-%d`, yy, mm, lastDay)

	rows, err := db.sql.Queryx(
		`SELECT * FROM transactions WHERE balance_id=$1 and date >= $2 and date <= $3 and delta < 0`,
		balanceID,
		compareStringBefore,
		compareStringUntil,
	)

	if err != nil {
		return nil, fmt.Errorf("queryx: %w", err)
	}

	type tr struct {
		UUID      string    `db:"uuid"`
		BalanceID int       `db:"balance_id"`
		Delta     float64   `db:"delta"`
		Date      time.Time `db:"date"`
		Tag       string    `db:"tag"`
	}

	result := make(map[string]float64)

	for rows.Next() {
		t := tr{}

		if err := rows.StructScan(&t); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		tag := t.Tag

		if tag == "" {
			tag = "others"
		}

		result[strings.ToLower(tag)] += math.Abs(t.Delta)
	}

	return result, nil
}
