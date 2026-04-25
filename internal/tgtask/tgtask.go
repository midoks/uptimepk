package tgtask

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"uptimepk/internal/log"

	"golang.org/x/net/proxy"

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

func GetBotByProxy(token, proxyURL string) (bot *tgbotapi.BotAPI, err error) {
	if proxyURL != "" {
		u, parseErr := url.Parse(proxyURL)
		if parseErr == nil {
			// 支持 HTTP/HTTPS/SOCKS5 代理
			if u.Scheme == "http" || u.Scheme == "https" {
				transport := &http.Transport{
					Proxy:           http.ProxyURL(u),
					TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
				}
				client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
				bot, err = tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, client)
			} else if u.Scheme == "socks5" {
				dialer, dialErr := proxy.FromURL(u, proxy.Direct)
				if dialErr != nil {
					log.Errorf("Failed to create SOCKS5 proxy dialer: %v", dialErr)
					bot, err = tgbotapi.NewBotAPI(token)
				} else {
					transport := &http.Transport{
						Dial:            dialer.Dial,
						TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
					}
					client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
					bot, err = tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, client)
				}
			} else {
				bot, err = tgbotapi.NewBotAPI(token)
			}

			if err != nil {
				// 只在首次错误时输出日志，避免日志刷屏
				if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "Client.Timeout") {
					log.Errorf("failed to create bot with proxy: %v", err)
				}
			}
		} else {
			bot, err = tgbotapi.NewBotAPI(token)
			if err != nil && !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "Client.Timeout") {
				log.Errorf("failed to create bot without proxy: %v", err)
			}
		}
	} else {
		bot, err = tgbotapi.NewBotAPI(token)
		if err != nil && !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "Client.Timeout") {
			log.Errorf("failed to create bot: %v", err)
		}
	}
	return bot, err
}

func (m *Manager) AddBot(id int64, token, proxyURL string, chatID int64, handler MessageHandler) error {
	m.mutex.Lock()

	// 检查是否存在相同 token 的 bot
	for existingID, existingBot := range m.bots {
		if existingBot.Token == token {
			// 先停止并删除旧的 bot
			if existingBot.running {
				close(existingBot.StopChan)
				existingBot.running = false
			}
			delete(m.bots, existingID)
			// fmt.Printf("Removed existing bot with same token: %d\n", existingID)
			// 等待足够的时间确保 Telegram 服务器释放旧的连接
			time.Sleep(5 * time.Second)
			break
		}
	}

	if _, exists := m.bots[id]; exists {
		m.mutex.Unlock()
		return fmt.Errorf("bot with id %d already exists", id)
	}

	m.mutex.Unlock()

	// 等待足够的时间确保所有 bot 实例完全停止
	time.Sleep(3 * time.Second)

	var err error
	var bot *tgbotapi.BotAPI
	bot, err = GetBotByProxy(token, proxyURL)
	if err != nil {
		log.Infof("creating bot with token: %s, proxy: %s", token, proxyURL)
		log.Errorf("creating bot error: %v", err)
		return err
	}

	// webhook deleted to avoid conflicts
	if err == nil && bot != nil {
		// 先删除 webhook，避免与其他实例冲突
		deleteConfig := tgbotapi.DeleteWebhookConfig{
			DropPendingUpdates: true,
		}
		_, _ = bot.Request(deleteConfig)

		// 使用同步的 getUpdates 清除旧的更新状态
		// 使用一个很大的 offset 值来跳过所有待处理的更新
		skipConfig := tgbotapi.NewUpdate(0)
		skipConfig.Timeout = 1
		updates, err := bot.GetUpdates(skipConfig)
		if err != nil {
			log.Errorf("failed to clear old updates: %v", err)
		} else {
			log.Infof("cleared %d old updates", len(updates))
		}
		// 等待足够的时间确保状态完全传播
		time.Sleep(2 * time.Second)
		// fmt.Println("cleared old updates state")
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
	//debug
	bot.Debug = false
	// fmt.Printf("Bot %d added: %s\n", id, bot.Self.UserName)
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
	// fmt.Printf("Bot %d started\n", id)
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
	// fmt.Printf("Bot %d stopped\n", id)
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
			bot.BotAPI.StopReceivingUpdates()
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			// fmt.Printf("收到来自 [%s] 的消息: %s\n", update.Message.From.UserName, update.Message.Text)
			if bot.MessageHandler != nil {
				if err := bot.MessageHandler(update, bot.BotAPI); err != nil {
					log.Errorf("处理消息失败: %v", err)
				}
			} else {
				// 默认处理逻辑
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyToMessageID = update.Message.MessageID

				if _, err := bot.BotAPI.Send(msg); err != nil {
					log.Errorf("发送消息失败: %v", err)
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
	// fmt.Printf("Bot %d removed\n", id)
	return nil
}

func (m *Manager) RemoveAllBots() error {
	m.mutex.Lock()

	for id, bot := range m.bots {
		if bot.running {
			close(bot.StopChan)
			bot.running = false
			// 等待一小段时间，确保 bot 完全停止
			time.Sleep(500 * time.Millisecond)
		}
		delete(m.bots, id)
		// fmt.Printf("Bot %d removed\n", id)
	}
	m.mutex.Unlock()
	// 等待所有 bot 实例完全停止
	time.Sleep(500 * time.Millisecond)
	return nil
}

func GetManager() *Manager {
	return botManager
}
