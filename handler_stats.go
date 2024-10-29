package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

func (p *processor) handlerStats(ctx context.Context, update tgbotapi.Update) ([]tgbotapi.MessageConfig, error) {
	bl, err := p.db.getBalance(ctx)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	now := time.Now()

	isNewDay := bl.UpdatedAt.Time.Day() != now.Day()
	if isNewDay {
		bl = startNewDayWithBalance(bl.Id, now, bl.Balance)
		bl, err = p.db.updateOnlyBalance(ctx, bl)
		if err != nil {
			return nil, fmt.Errorf("update balance: %w", err)
		}
	}

	msg := fmt.Sprintf(
		"balance: %.2f\n"+
			"today: %.2f\n"+
			"day limit: %.2f",
		bl.Balance,
		bl.Status,
		bl.DayLimit,
	)

	daysLeft := monthLastDay(now) - now.Day() + 1

	tomorrowLimit := bl.Balance
	if daysLeft > 1 {
		tomorrowLimit = bl.Balance / float64(daysLeft-1)
	}

	msg = fmt.Sprintf(
		"%s\n"+
			"tomorrow limit: %.2f",
		msg,
		tomorrowLimit)

	msg = fmt.Sprintf(
		"%s\n"+
			"days left: %d",
		msg,
		daysLeft)

	if bl.TodaySpent > 0 {
		msg = fmt.Sprintf(
			"%s\n"+
				"today spent: %.2f",
			msg,
			bl.TodaySpent)
	}

	if bl.TodayAdded > 0 {
		msg = fmt.Sprintf(
			"%s\n"+
				"today added: %.2f",
			msg,
			bl.TodayAdded)
	}

	tgMsg := tgbotapi.NewMessage(update.Message.From.ID, msg)
	return asSlice(tgMsg), nil
}
