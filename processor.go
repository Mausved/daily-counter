package main

import (
	"context"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"time"
)

type processor struct {
	db *Database
}

func (p *processor) process(ctx context.Context, update tgbotapi.Update) (messages []tgbotapi.MessageConfig) {
	if update.Message == nil {
		return nil
	}

	if update.Message.From.ID != yin && update.Message.From.ID != yang {
		return nil
	}

	errMsg := func() tgbotapi.MessageConfig {
		m := tgbotapi.NewMessage(update.Message.From.ID, "ошибОчка, коряво написал, фраер")
		m.ReplyToMessageID = update.Message.MessageID
		return m
	}

	defer func() {
		if err := recover(); err != nil {
			messages = asSlice(errMsg())
		}
	}()

	text := update.Message.Text

	switch {
	case plusPattern.MatchString(text):
		messages, err := p.handlerPlus(ctx, update)
		if err != nil {
			log.Printf("failed handler plus: %v", err)
			return asSlice(errMsg())
		}
		return messages
	case minusPattern.MatchString(text):
		messages, err := p.handlerMinus(ctx, update)
		if err != nil {
			log.Printf("failed handler minus: %v", err)
			return asSlice(errMsg())
		}
		return messages
	case setBalancePattern.MatchString(text):
		messages, err := p.handlerSetBalance(ctx, update)
		if err != nil {
			log.Printf("failed handler set balance: %v", err)
			return asSlice(errMsg())
		}
		return messages
	case strings.ToLower(text) == "s":
		messages, err := p.handlerStats(ctx, update)
		if err != nil {
			log.Printf("failed handler stats: %v", err)
			return asSlice(errMsg())
		}
		return messages
	default:
		msg := "ну ты совсем куку...\n" +
			"вот тебе команды, балда:\n" +
			"+x comment - add income\n" +
			"-x comment - add consumption\n" +
			"bXXX comment - set balance=XXX\n" +
			"s - get statistics\n"
		tgMsg := tgbotapi.NewMessage(update.Message.From.ID, msg)
		return asSlice(tgMsg)
	}

}

func startNewDayWithBalance(startTime time.Time, balance float64) *balanceLimit {
	limit := countDayLimit(balance)

	bl := &balanceLimit{
		Balance:  balance,
		Status:   limit,
		DayLimit: limit,
		UpdatedAt: sql.NullTime{
			Time:  startTime,
			Valid: true,
		},
		TodaySpent: 0,
		TodayAdded: 0,
	}

	return bl
}

func asSlice(msg ...tgbotapi.MessageConfig) []tgbotapi.MessageConfig {
	return msg
}

func countDayLimit(balance float64) float64 {
	now := time.Now()

	lastDay := monthLastDay(now)
	daysLeft := lastDay - now.Day() + 1
	dailyConsumptionLimit := balance / float64(daysLeft)

	return dailyConsumptionLimit
}

func monthLastDay(t time.Time) int {
	for i := 27; ; i++ {
		nextDayDate := time.Date(t.Year(), t.Month(), i, 0, 0, 0, 0, t.Location())
		if nextDayDate.Month() != t.Month() {
			return i - 1
		}
	}
}

func valueFromMessageText(text string) (float64, error) {
	separateBySpaces := strings.Fields(text)

	if len(separateBySpaces) == 0 {
		return 0, fmt.Errorf("invalid format")
	}

	value := separateBySpaces[0]

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse float: %w", err)
	}

	return parsed, nil
}
