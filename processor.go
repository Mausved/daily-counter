package main

import (
	"context"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var balancePattern = regexp.MustCompile(`^b\d*`)
var plusPattern = regexp.MustCompile(`^\+\d*`)
var minusPattern = regexp.MustCompile(`^-\d*`)

type processor struct {
	db *Database
}

func (p *processor) process(ctx context.Context, update tgbotapi.Update) []tgbotapi.MessageConfig {
	if update.Message == nil {
		return nil
	}

	if update.Message.From.ID != yin && update.Message.From.ID != yang {
		return nil
	}

	text := update.Message.Text

	errMsg := func() tgbotapi.MessageConfig {
		m := tgbotapi.NewMessage(update.Message.From.ID, "ошибОчка, коряво написал, фраер")
		m.ReplyToMessageID = update.Message.MessageID
		return m
	}

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
	case balancePattern.MatchString(text):
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

func (p *processor) handlerPlus(ctx context.Context, update tgbotapi.Update) ([]tgbotapi.MessageConfig, error) {
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
	bl.Balance += parsed

	bl.Status += parsed
	bl.TodayAdded += parsed

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

func (p *processor) handlerMinus(ctx context.Context, update tgbotapi.Update) ([]tgbotapi.MessageConfig, error) {
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

func (p *processor) handlerSetBalance(ctx context.Context, update tgbotapi.Update) ([]tgbotapi.MessageConfig, error) {
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

func (p *processor) handlerStats(ctx context.Context, update tgbotapi.Update) ([]tgbotapi.MessageConfig, error) {
	bl, err := p.db.getBalance(ctx)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	now := time.Now()

	isNewDay := bl.UpdatedAt.Time.Day() != now.Day()
	if isNewDay {
		bl = startNewDayWithBalance(now, bl.Balance)
		bl, err = p.db.updateBalance(ctx, bl)
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
	value := separateBySpaces[0]

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse float: %w", err)
	}

	return parsed, nil
}
