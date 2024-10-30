package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
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
	yin  = 741126351
	yang = 381523363
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	viper.AutomaticEnv()

	tgBotToken := viper.GetString("TELEGRAM_BOT_API_TOKEN")
	if tgBotToken == "" {
		log.Fatalf("empty telegram bot api token")
	}

	dbConn := viper.GetString("POSTGRES_DSN")
	if dbConn == "" {
		log.Fatalf("empty db conn string")
	}

	db, err := initDatabase(dbConn)
	if err != nil {
		log.Fatalf("failed init database: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(tgBotToken)
	if err != nil {
		log.Fatalf("failed new bot api: %v", err)
	}

	bot.Debug = appMode == appModeDev

	if !bot.Debug {
		if err := sendHello(bot); err != nil {
			log.Fatalf("failed send hello message: %v", err)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := start(ctx, bot, db); err != nil {
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

func start(ctx context.Context, bot *tgbotapi.BotAPI, db *Database) error {
	p := &processor{db: db}

	updates := bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updates:
			func() {
				ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
				defer cancel()

				messagesToSent := p.process(ctx, update)
				for _, msg := range messagesToSent {
					if _, err := bot.Send(msg); err != nil {
						log.Printf("failed send msg: %v\n", err)
						continue
					}
				}
			}()
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
