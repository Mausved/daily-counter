package main

import (
	"context"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math"
	"regexp"
	"time"
)

var minusPattern = regexp.MustCompile(`^-\d*`)

func (p *processor) handlerMinus(ctx context.Context, update tgbotapi.Update) ([]tgbotapi.MessageConfig, error) {
	if len(update.Message.Text) < 1 {
		return nil, fmt.Errorf("invalid set balance pattern")
	}

	bl, err := p.db.getBalance(ctx)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	now := time.Now()

	isNewDay := bl.UpdatedAt.Time.Day() != now.Day()
	if isNewDay {
		bl = startNewDayWithBalance(now, bl.Balance)
	}

	parsed, err := valueFromMessageText(update.Message.Text)
	if err != nil {
		return nil, err
	}

	parsed = math.Abs(parsed)
	bl.Balance -= parsed
	if bl.Balance < 0 {
		bl.Balance = 0
	}

	bl.TodaySpent += parsed
	bl.Status -= parsed
	if bl.Balance == 0 {
		bl.Status = 0
	}

	bl.UpdatedAt = sql.NullTime{
		Time:  now,
		Valid: true,
	}

	updated, err := p.db.updateBalance(ctx, bl)
	if err != nil {
		return nil, fmt.Errorf("update balance: %w", err)
	}

	msg := fmt.Sprintf("today left: %.2f", updated.Status)

	msgYin := tgbotapi.NewMessage(yin, msg)
	msgYang := tgbotapi.NewMessage(yang, msg)

	if update.Message.From.ID == yin {
		msgYang.Text = fmt.Sprintf("updated by @%s\n%s\n%s", update.Message.From.UserName, update.Message.Text, msgYang.Text)
		msgYin.ReplyToMessageID = update.Message.MessageID
	}

	if update.Message.From.ID == yang {
		msgYin.Text = fmt.Sprintf("updated by @%s\n%s\n%s", update.Message.From.UserName, update.Message.Text, msgYin.Text)
		msgYang.ReplyToMessageID = update.Message.MessageID
	}

	return asSlice(msgYin, msgYang), nil
}
