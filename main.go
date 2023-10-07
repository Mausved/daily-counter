package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	appModeDev  = "dev"
	appModeProd = "prod"
)

const appMode = appModeProd

const (
	fileName   = "balance.daily"
	yin        = 741126351
	yang       = 381523363
	tgBotToken = "6628221972:AAHJLliOWzvLMN5Fwfqu5kiGgehQyc4vh-0"
)

func main() {
	balanceLimit, err := readBalanceFromFile(fileName)
	if err != nil {
		log.Fatalf("failed read balance from file: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(tgBotToken)
	if err != nil {
		log.Fatalf("failed new bot api: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	bot.Debug = appMode == appModeDev

	if !bot.Debug {
		if err := sendHello(bot); err != nil {
			log.Fatalf("failed send hello message: %v", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := start(ctx, bot, balanceLimit); err != nil {
			log.Printf("failed start: %v", err)
		}
	}()

	gracefulShutdown(cancel, &wg)
	if !bot.Debug {
		if err := sendGoodbye(bot); err != nil {
			log.Fatalf("failed send goodbye message: %v", err)
		}
	}
}

func readBalanceFromFile(fileName string) (*balanceLimit, error) {
	b, err := os.ReadFile(fileName)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &balanceLimit{}, nil
		}
		return nil, fmt.Errorf("open file %s, err: %w", fileName, err)
	}

	l := &balanceLimit{}
	if err := json.Unmarshal(b, &l); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	if l == nil {
		return &balanceLimit{}, nil
	}

	return l, nil
}

func writeBalanceToFile(bl *balanceLimit, fileName string) error {
	b, err := json.Marshal(bl)
	if err != nil {
		return fmt.Errorf("unmarshal balance: %w", err)
	}

	if err := os.WriteFile(fileName, b, 0666); err != nil {
		return fmt.Errorf("ailed write to file: %w", err)
	}
	return nil
}

func sendHello(bot *tgbotapi.BotAPI) error {
	msgYin := tgbotapi.NewMessage(yin, "привет, Фродо")
	if _, err := bot.Send(msgYin); err != nil {
		return fmt.Errorf("send message to yin: %w", err)
	}

	msgYang := tgbotapi.NewMessage(yang, "привет, Бильбо")
	if _, err := bot.Send(msgYang); err != nil {
		return fmt.Errorf("send message to yang: %w", err)
	}

	return nil
}

func sendGoodbye(bot *tgbotapi.BotAPI) error {
	msgYin := tgbotapi.NewMessage(yin, "пока, Фродо")
	if _, err := bot.Send(msgYin); err != nil {
		return fmt.Errorf("send message to yin: %w", err)
	}

	msgYang := tgbotapi.NewMessage(yang, "пока, Бильбо")
	if _, err := bot.Send(msgYang); err != nil {
		return fmt.Errorf("send message to yang: %w", err)
	}

	return nil
}

func start(ctx context.Context, bot *tgbotapi.BotAPI, balanceLimit *balanceLimit) error {
	balanceSaveTicker := time.NewTicker(5 * time.Minute)
	p := &processor{bl: balanceLimit}
	defer func() {
		if err := writeBalanceToFile(p.bl, fileName); err != nil {
			log.Panic("write balance to file")
		}
	}()

	updates := bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-balanceSaveTicker.C:
			if err := writeBalanceToFile(p.bl, fileName); err != nil {
				log.Printf("failed write balance to file: %v", err)
			}
		case update := <-updates:
			messagesToSent := p.process(update)
			for _, msg := range messagesToSent {
				if _, err := bot.Send(msg); err != nil {
					log.Printf("failed send msg: %v\n", err)
					continue
				}
			}
		}
	}
}

func gracefulShutdown(cancel func(), wg *sync.WaitGroup) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch
	signal.Stop(ch)
	cancel()
	wg.Wait()
}

type balanceLimit struct {
	Balance    float64
	Status     float64
	DayLimit   float64
	UpdateAt   time.Time
	TodaySpent float64
	TodayAdded float64
}
