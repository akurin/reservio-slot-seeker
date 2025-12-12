package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const pollInterval = 1 * time.Second

type service struct {
	name string
	id   string
}

var services = []service{
	{
		name: "citizenship",
		id:   "63fe0e8c-b127-43e3-874a-bac9c660045b",
	},
}

const reservioURLTemplate = "https://ambasada-r-moldova-in-f-rusa.reservio.com/api/v2/businesses/" +
	"09250556-2450-437f-aede-82e78712f114/" +
	"availability/booking-days?" +
	"filter[from]=%s&" +
	"filter[resourceId]=&" +
	"filter[serviceId]=%s&" +
	"filter[to]=%s&" +
	"ignoreBookingBoundaries=0"

var from = time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC)
var to = time.Date(2026, 01, 01, 0, 0, 0, 0, time.UTC)

var maxRandomOffset = 24 * time.Hour

type reservioResponse struct {
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

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN env variable not set")
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

	// Start monitoring each service
	for _, svc := range services {
		if svc.id == "" {
			log.Printf("Skipping service '%s' - no ID provided", svc.name)
			continue
		}
		go monitorService(svc, bot, chatID)
	}

	// Keep the main goroutine alive
	select {}
}

func monitorService(svc service, bot *tgbotapi.BotAPI, chatID int64) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	var previousAvailableDays []string

	log.Printf("Starting monitoring for service: %s (ID: %s)", svc.name, svc.id)

	for {
		select {
		case <-ticker.C:
			result, err := getSlots(svc)
			if err != nil {
				log.Printf("[%s] Error getting slots: %v", svc.name, err)
				continue
			}

			log.Printf("[%s] Total: %d", svc.name, result.Meta.Total)

			if len(result.Data) > 0 {
				log.Printf("[%s] Date range: %s - %s", svc.name, result.Data[0].Attributes.Date, result.Data[len(result.Data)-1].Attributes.Date)
			}

			var availableDays []string
			for _, day := range result.Data {
				if day.Attributes.IsAvailable {
					availableDays = append(availableDays, day.Attributes.Date)
				}
			}

			if !slicesEqual(availableDays, previousAvailableDays) {
				var msgText string
				if len(availableDays) > 0 {
					msgText = fmt.Sprintf("🎯 [%s] Available days: %s", svc.name, strings.Join(availableDays, ", "))
				} else {
					msgText = fmt.Sprintf("❌ [%s] No available days found", svc.name)
				}

				msg := tgbotapi.NewMessage(chatID, msgText)
				_, err = bot.Send(msg)
				if err != nil {
					log.Printf("[%s] Error sending Telegram message: %v", svc.name, err)
				} else {
					log.Printf("[%s] Sent notification: %s", svc.name, msgText)
				}

				previousAvailableDays = availableDays
			} else {
				log.Printf("[%s] No change in available days (count: %d)", svc.name, len(availableDays))
			}
		}
	}
}

func getSlots(svc service) (reservioResponse, error) {
	// Add random offset to bypass cache
	fromWithRandomOffset := from.Add(randomOffset())
	toWithRandomOffset := to.Add(randomOffset())

	reservioURL := fmt.Sprintf(reservioURLTemplate,
		formatDate(fromWithRandomOffset),
		svc.id,
		formatDate(toWithRandomOffset),
	)

	log.Printf("[%s] Fetching slots from: %s", svc.name, reservioURL)

	req, err := http.NewRequest(http.MethodGet, reservioURL, nil)
	if err != nil {
		return reservioResponse{}, fmt.Errorf("failed to create Reservio request: %w", err)
	}

	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return reservioResponse{}, fmt.Errorf("failed to make Reservio request: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	var result reservioResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return reservioResponse{}, fmt.Errorf("failed to decode Reservio response: %w", err)
	}

	return result, nil
}

func randomOffset() time.Duration {
	return time.Duration(rand.Int63n(maxRandomOffset.Nanoseconds()))
}

func formatDate(date time.Time) string {
	return date.Format("2006-01-02T15:04:05.000Z")
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
