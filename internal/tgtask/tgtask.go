package tgtask

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	Token    string
	Proxy    string
	ChatID   int64
	BotAPI   *tgbotapi.BotAPI
	StopChan chan struct{}
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

func (m *Manager) AddBot(id int64, token, proxy string, chatID int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.bots[id]; exists {
		return fmt.Errorf("bot with id %d already exists", id)
	}

	var bot *tgbotapi.BotAPI
	var err error
	if proxy != "" {
		u, parseErr := url.Parse(proxy)
		if parseErr == nil {
			tr := &http.Transport{Proxy: http.ProxyURL(u)}
			client := &http.Client{Transport: tr}
			bot, err = tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, client)
		} else {
			bot, err = tgbotapi.NewBotAPI(token)
		}
	} else {
		bot, err = tgbotapi.NewBotAPI(token)
	}

	if err != nil {
		return fmt.Errorf("failed to create bot: %v", err)
	}

	newBot := &Bot{
		Token:    token,
		Proxy:    proxy,
		ChatID:   chatID,
		BotAPI:   bot,
		StopChan: make(chan struct{}),
	}

	m.bots[id] = newBot
	bot.Debug = true
	fmt.Printf("Bot %d added: %s\n", id, bot.Self.UserName)

	return nil
}

func (m *Manager) StartBot(id int64) error {
	m.mutex.RLock()
	bot, exists := m.bots[id]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("bot with id %d not found", id)
	}

	select {
	case <-bot.StopChan:
	default:
		return fmt.Errorf("bot with id %d is already running", id)
	}

	bot.StopChan = make(chan struct{})
	go m.runBot(bot)
	fmt.Printf("Bot %d started\n", id)
	return nil
}

func (m *Manager) StopBot(id int64) error {
	m.mutex.RLock()
	bot, exists := m.bots[id]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("bot with id %d not found", id)
	}

	select {
	case <-bot.StopChan:
		return fmt.Errorf("bot with id %d is not running", id)
	default:
		close(bot.StopChan)
	}

	fmt.Printf("Bot %d stopped\n", id)
	return nil
}

func (m *Manager) runBot(bot *Bot) {
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

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			if _, err := bot.BotAPI.Send(msg); err != nil {
				fmt.Printf("发送消息失败: %v\n", err)
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

	select {
	case <-bot.StopChan:
	default:
		close(bot.StopChan)
	}

	delete(m.bots, id)
	fmt.Printf("Bot %d removed\n", id)
	return nil
}

func GetManager() *Manager {
	return botManager
}
