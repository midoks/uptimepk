package tgtask

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageHandler func(update tgbotapi.Update, bot *tgbotapi.BotAPI) error

type Bot struct {
	Token          string
	Proxy          string
	ChatID         int64
	BotAPI         *tgbotapi.BotAPI
	StopChan       chan struct{}
	running        bool
	MessageHandler MessageHandler
}

type Manager struct {
	bots  map[int64]*Bot
	mutex sync.RWMutex
}

var botManager *Manager

func init() {
	botManager = &Manager{
		bots: make(map[int64]*Bot),
	}
}

func (m *Manager) AddBot(id int64, token, proxyURL string, chatID int64, handler MessageHandler) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.bots[id]; exists {
		return fmt.Errorf("bot with id %d already exists", id)
	}

	fmt.Printf("Creating bot with token: %s, proxy: %s\n", token, proxyURL)

	var bot *tgbotapi.BotAPI
	var err error

	if proxyURL != "" {
		u, parseErr := url.Parse(proxyURL)
		if parseErr == nil {
			fmt.Printf("Using proxy: %s\n", u.String())

			// 支持 HTTP/HTTPS 代理
			if u.Scheme == "http" || u.Scheme == "https" {
				transport := &http.Transport{
					Proxy:           http.ProxyURL(u),
					TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
				}
				client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
				bot, err = tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, client)
			} else {
				fmt.Printf("Unsupported proxy scheme: %s\n", u.Scheme)
				bot, err = tgbotapi.NewBotAPI(token)
			}

			if err != nil {
				fmt.Printf("Failed to create bot with proxy: %v\n", err)
			}
		} else {
			fmt.Printf("Failed to parse proxy URL: %v\n", parseErr)
			bot, err = tgbotapi.NewBotAPI(token)
			if err != nil {
				fmt.Printf("Failed to create bot without proxy: %v\n", err)
			}
		}
	} else {
		fmt.Println("No proxy specified, creating bot directly")
		bot, err = tgbotapi.NewBotAPI(token)
		if err != nil {
			fmt.Printf("Failed to create bot: %v\n", err)
		}
	}

	if err != nil || bot == nil {
		return fmt.Errorf("failed to create bot: %v", err)
	}

	newBot := &Bot{
		Token:          token,
		Proxy:          proxyURL,
		ChatID:         chatID,
		BotAPI:         bot,
		StopChan:       make(chan struct{}),
		MessageHandler: handler,
	}

	m.bots[id] = newBot
	bot.Debug = true
	fmt.Printf("Bot %d added: %s\n", id, bot.Self.UserName)

	return nil
}

func (m *Manager) StartBot(id int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	bot, exists := m.bots[id]
	if !exists {
		return fmt.Errorf("bot with id %d not found", id)
	}

	if bot.running {
		return fmt.Errorf("bot with id %d is already running", id)
	}

	bot.StopChan = make(chan struct{})
	bot.running = true
	go m.runBot(bot)
	fmt.Printf("Bot %d started\n", id)
	return nil
}

func (m *Manager) StopBot(id int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	bot, exists := m.bots[id]
	if !exists {
		return fmt.Errorf("bot with id %d not found", id)
	}

	if !bot.running {
		return fmt.Errorf("bot with id %d is not running", id)
	}

	close(bot.StopChan)
	bot.running = false
	fmt.Printf("Bot %d stopped\n", id)
	return nil
}

func (m *Manager) runBot(bot *Bot) {
	defer func() {
		bot.running = false
	}()

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.BotAPI.GetUpdatesChan(updateConfig)

	for {
		select {
		case <-bot.StopChan:
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			fmt.Printf("收到来自 [%s] 的消息: %s\n", update.Message.From.UserName, update.Message.Text)

			if bot.MessageHandler != nil {
				if err := bot.MessageHandler(update, bot.BotAPI); err != nil {
					fmt.Printf("处理消息失败: %v\n", err)
				}
			} else {
				// 默认处理逻辑
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyToMessageID = update.Message.MessageID

				if _, err := bot.BotAPI.Send(msg); err != nil {
					fmt.Printf("发送消息失败: %v\n", err)
				}
			}
		}
	}
}

func (m *Manager) RemoveBot(id int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	bot, exists := m.bots[id]
	if !exists {
		return fmt.Errorf("bot with id %d not found", id)
	}

	if bot.running {
		close(bot.StopChan)
		bot.running = false
	}

	delete(m.bots, id)
	fmt.Printf("Bot %d removed\n", id)
	return nil
}

func (m *Manager) RemoveAllBots() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for id, bot := range m.bots {
		if bot.running {
			close(bot.StopChan)
			bot.running = false
		}
		delete(m.bots, id)
		fmt.Printf("Bot %d removed\n", id)
	}

	return nil
}

func GetManager() *Manager {
	return botManager
}
