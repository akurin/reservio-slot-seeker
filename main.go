package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN env variable not set")
	}

	reservioURL := os.Getenv("RESERVIO_URL")
	if reservioURL == "" {
		log.Fatal("RESERVIO_URL env variable not set")
	}

	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		log.Fatal("TELEGRAM_CHAT_ID env variable not set")
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid TELEGRAM_CHAT_ID: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)
	go func() {
		for update := range updates {
			if update.Message != nil {
				log.Printf("Received message from chat ID: %d, user: %s", update.Message.Chat.ID, update.Message.From.UserName)
			}
		}
	}()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			resp, err := http.Get(reservioURL)
			if err != nil {
				log.Println("Error fetching Reservio URL:", err)
				continue
			}
			defer resp.Body.Close()

			var result struct {
				Meta struct {
					Total int `json:"total"`
				} `json:"meta"`
				Data []struct {
					Attributes struct {
						Date        string `json:"date"`
						IsAvailable bool   `json:"isAvailable"`
					} `json:"attributes"`
				} `json:"data"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				log.Println("Error decoding Reservio response:", err)
				continue
			}

			log.Printf("Total: %d", result.Meta.Total)

			var availableDays []string
			for _, day := range result.Data {
				if day.Attributes.IsAvailable {
					availableDays = append(availableDays, day.Attributes.Date)
				}
			}

			if len(availableDays) > 0 {
				msgText := "Available days: " + strings.Join(availableDays, ", ")
				msg := tgbotapi.NewMessage(chatID, msgText)
				_, err = bot.Send(msg)
				if err != nil {
					log.Println("Error sending Telegram message:", err)
				}
			}
		}
	}
}
