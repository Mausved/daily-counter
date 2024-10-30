package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sort"
)

func (p *processor) handlerFullTagStats(ctx context.Context, update tgbotapi.Update) ([]tgbotapi.MessageConfig, error) {
	bl, err := p.db.getBalance(ctx)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	result, err := p.db.getSpendingTransactionsForMonth(ctx, int(bl.Id))
	if err != nil {
		return nil, fmt.Errorf("get transactions")
	}

	pairs := make([]struct {
		key   string
		value float64
	}, 0, len(result))

	for k, v := range result {
		pairs = append(pairs, struct {
			key   string
			value float64
		}{key: k, value: v})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].value > pairs[j].value
	})

	msg := "stat spending"

	if len(pairs) == 0 {
		msg = fmt.Sprintf("%s: nothing!", msg)
	} else {
		msg = fmt.Sprintf("%s:\n", msg)
	}

	for _, pair := range pairs {
		msg = fmt.Sprintf("%s\n- %s: %.2f", msg, pair.key, pair.value)
	}

	tgMsg := tgbotapi.NewMessage(update.Message.From.ID, msg)
	return asSlice(tgMsg), nil
}
