package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math"
	"regexp"
	"time"
)

var setBalancePattern = regexp.MustCompile(`^b\d*`)

func (p *processor) handlerSetBalance(ctx context.Context, update tgbotapi.Update) ([]tgbotapi.MessageConfig, error) {
	if len(update.Message.Text) < 1 {
		return nil, fmt.Errorf("invalid set balance pattern")
	}

	balance, err := valueFromMessageText(update.Message.Text[1:])
	if err != nil {
		return nil, err
	}

	balance = math.Abs(balance)
	bl := startNewDayWithBalance(time.Now(), balance)

	updated, err := p.db.updateBalance(ctx, bl)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	msg := fmt.Sprintf("set balance: %.2f", updated.Balance)

	msgYin := tgbotapi.NewMessage(yin, msg)
	msgYang := tgbotapi.NewMessage(yang, msg)

	if update.Message.From.ID == yin {
		msgYang.Text = fmt.Sprintf("setted balance by @%s\n%s\n%s", update.Message.From.UserName, update.Message.Text, msgYang.Text)
		msgYin.ReplyToMessageID = update.Message.MessageID
	}

	if update.Message.From.ID == yang {
		msgYin.Text = fmt.Sprintf("setted balance by @%s\n%s\n%s", update.Message.From.UserName, update.Message.Text, msgYin.Text)
		msgYang.ReplyToMessageID = update.Message.MessageID
	}

	return asSlice(msgYin, msgYang), nil
}
