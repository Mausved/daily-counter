package main

import (
	"database/sql"
	"time"
)

type transaction struct {
	UUID      string    `db:"uuid"`
	BalanceID int       `db:"balance_id"`
	Delta     float64   `db:"delta"`
	Date      time.Time `db:"date"`
	Tag       string    `db:"tag"`
}

type balanceLimit struct {
	Id         int64        `db:"id"`
	Balance    float64      `db:"balance"`
	Status     float64      `db:"status"`
	DayLimit   float64      `db:"day_limit"`
	UpdatedAt  sql.NullTime `db:"updated_at"`
	TodaySpent float64      `db:"today_spent"`
	TodayAdded float64      `db:"today_added"`
	Name       string       `db:"name"`
}
