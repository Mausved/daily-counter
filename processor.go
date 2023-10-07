package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var balancePattern = regexp.MustCompile(`b\d*`)
var plusPattern = regexp.MustCompile(`\+\d*`)
var minusPattern = regexp.MustCompile(`-\d*`)

type processor struct {
	bl *balanceLimit
}

func (p *processor) process(update tgbotapi.Update) []tgbotapi.MessageConfig {
	if update.Message == nil { // If we got a message
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
		messages, err := p.handlerPlus(update)
		if err != nil {
			return asSlice(errMsg())
		}
		return messages
	case minusPattern.MatchString(text):
		messages, err := p.handlerMinus(update)
		if err != nil {
			return asSlice(errMsg())
		}
		return messages
	case balancePattern.MatchString(text):
		messages, err := p.handlerSetBalance(update)
		if err != nil {
			return asSlice(errMsg())
		}
		return messages
	case strings.ToLower(text) == "s":
		messages, err := p.handlerStats(update)
		if err != nil {
			return asSlice(errMsg())
		}
		return messages
	default:
		msg := "ну ты совсем куку...\n" +
			"вот тебе команды, балда:\n" +
			"+x - add income\n" +
			"-x - add consumption\n" +
			"bXXX - set balance=XXX\n" +
			"s - get statistics\n"
		tgMsg := tgbotapi.NewMessage(update.Message.From.ID, msg)
		return asSlice(tgMsg)
	}

}

func (p *processor) handlerPlus(update tgbotapi.Update) ([]tgbotapi.MessageConfig, error) {
	text := update.Message.Text
	parsed, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return nil, fmt.Errorf("parse float: %w", err)
	}

	parsed = math.Abs(parsed)
	p.bl.Balance += parsed

	now := time.Now()

	isNewDay := p.bl.UpdateAt.Day() != now.Day()
	if isNewDay {
		p.startNewDay(now)
	}

	p.bl.Status += parsed
	p.bl.TodayAdded += parsed

	msg := fmt.Sprintf("today: %.2f", p.bl.Status)

	msgYin := tgbotapi.NewMessage(yin, msg)
	msgYang := tgbotapi.NewMessage(yang, msg)

	if update.Message.From.ID == yin {
		msgYang.Text = fmt.Sprintf("updated by @%s\n%s\n%s", update.Message.From.UserName, text, msgYang.Text)
		msgYin.ReplyToMessageID = update.Message.MessageID
	}

	if update.Message.From.ID == yang {
		msgYin.Text = fmt.Sprintf("updated by @%s\n%s\n%s", update.Message.From.UserName, text, msgYin.Text)
		msgYang.ReplyToMessageID = update.Message.MessageID
	}

	return asSlice(msgYin, msgYang), nil
}

func (p *processor) actualizeStats() {
	now := time.Now()
	if p.bl.UpdateAt.Day() == now.Day() {
		return
	}

	p.bl.Status = countDayLimit(p.bl.Balance)
	p.bl.DayLimit = p.bl.Status
	p.bl.UpdateAt = now
}

func (p *processor) handlerMinus(update tgbotapi.Update) ([]tgbotapi.MessageConfig, error) {
	text := update.Message.Text

	parsed, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return nil, fmt.Errorf("parse float: %w", err)
	}

	parsed = math.Abs(parsed)
	p.bl.Balance -= parsed
	if p.bl.Balance < 0 {
		p.bl.Balance = 0
	}

	now := time.Now()

	isNewDay := p.bl.UpdateAt.Day() != now.Day()
	if isNewDay {
		p.startNewDay(now)
	}

	p.bl.TodaySpent += parsed
	p.bl.Status -= parsed
	if p.bl.Balance == 0 {
		p.bl.Status = 0
	}

	p.bl.UpdateAt = now

	msg := fmt.Sprintf("today: %.2f", p.bl.Status)

	msgYin := tgbotapi.NewMessage(yin, msg)
	msgYang := tgbotapi.NewMessage(yang, msg)

	if update.Message.From.ID == yin {
		msgYang.Text = fmt.Sprintf("updated by @%s\n%s\n%s", update.Message.From.UserName, text, msgYang.Text)
		msgYin.ReplyToMessageID = update.Message.MessageID
	}

	if update.Message.From.ID == yang {
		msgYin.Text = fmt.Sprintf("updated by @%s\n%s\n%s", update.Message.From.UserName, text, msgYin.Text)
		msgYang.ReplyToMessageID = update.Message.MessageID
	}

	return asSlice(msgYin, msgYang), nil
}

func (p *processor) handlerSetBalance(update tgbotapi.Update) ([]tgbotapi.MessageConfig, error) {
	text := update.Message.Text

	parsed, err := strconv.ParseFloat(text[1:], 64)
	if err != nil {
		return nil, fmt.Errorf("parse float: %w", err)
	}

	parsed = math.Abs(parsed)

	dayLimit := countDayLimit(parsed)

	p.bl.Balance = parsed
	p.bl.Status = dayLimit
	p.bl.DayLimit = dayLimit
	p.bl.TodayAdded = 0
	p.bl.TodaySpent = 0

	msg := fmt.Sprintf("set balance: %.2f", p.bl.Balance)

	msgYin := tgbotapi.NewMessage(yin, msg)
	msgYang := tgbotapi.NewMessage(yang, msg)

	if update.Message.From.ID == yin {
		msgYang.Text = fmt.Sprintf("setted balance by @%s\n%s\n%s", update.Message.From.UserName, text, msgYang.Text)
		msgYin.ReplyToMessageID = update.Message.MessageID
	}

	if update.Message.From.ID == yang {
		msgYin.Text = fmt.Sprintf("setted balance by @%s\n%s\n%s", update.Message.From.UserName, text, msgYin.Text)
		msgYang.ReplyToMessageID = update.Message.MessageID
	}

	return asSlice(msgYin, msgYang), nil
}

func (p *processor) handlerStats(update tgbotapi.Update) ([]tgbotapi.MessageConfig, error) {
	now := time.Now()

	isNewDay := p.bl.UpdateAt.Day() != now.Day()
	if isNewDay {
		p.startNewDay(now)
	}

	msg := fmt.Sprintf(
		"balance: %.2f\n"+
			"status: %.2f\n"+
			"day limit: %.2f",
		p.bl.Balance,
		p.bl.Status,
		p.bl.DayLimit,
	)

	if p.bl.TodaySpent > 0 {
		msg = fmt.Sprintf(
			"%s\n"+
				"today spent: %.2f",
			msg,
			p.bl.TodaySpent)
	}

	if p.bl.TodayAdded > 0 {
		msg = fmt.Sprintf(
			"%s\n"+
				"today added: %.2f",
			msg,
			p.bl.TodayAdded)
	}

	tgMsg := tgbotapi.NewMessage(update.Message.From.ID, msg)
	return asSlice(tgMsg), nil
}

func asSlice(msg ...tgbotapi.MessageConfig) []tgbotapi.MessageConfig {
	return msg
}

func countDayLimit(balance float64) float64 {
	now := time.Now()

	y := now.Year()
	m := now.Month()
	d := now.Day()

	lastDay := 0
	for i := 27; ; i++ {
		nextDayDate := time.Date(y, m, i, 0, 0, 0, 0, now.Location())
		if nextDayDate.Month() != m {
			lastDay = i - 1
			break
		}
	}

	daysLeft := lastDay - d + 1
	dailyConsumptionLimit := balance / float64(daysLeft)

	return dailyConsumptionLimit
}

func (p *processor) startNewDay(now time.Time) {
	limit := countDayLimit(p.bl.Balance)
	p.bl.Status = limit
	p.bl.DayLimit = limit
	p.bl.UpdateAt = now
	p.bl.TodaySpent = 0
	p.bl.TodayAdded = 0
}
