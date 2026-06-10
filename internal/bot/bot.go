// Package bot wires the Telegram bot to the queue and player.
package bot

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"music-queue/internal/player"
	"music-queue/internal/queue"
	"music-queue/internal/ytinfo"
)

// ytURL matches youtube.com, music.youtube.com, and youtu.be links.
var ytURL = regexp.MustCompile(`https?://(?:www\.|music\.|m\.)?(?:youtube\.com/\S+|youtu\.be/\S+)`)

// Bot couples the Telegram API client with the queue and player.
type Bot struct {
	api        *tgbotapi.BotAPI
	q          *queue.Queue
	player     *player.Player
	passphrase string
	mu         sync.RWMutex
	authorized map[int64]bool
}

// New creates a Bot from a token and an optional passphrase. It also wires
// player notifications back to Telegram so playback updates are pushed to the
// requesting chat.
func New(token, passphrase string, q *queue.Queue, p *player.Player) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	b := &Bot{api: api, q: q, player: p, passphrase: passphrase, authorized: map[int64]bool{}}

	p.Notify = func(chatID int64, text string) {
		if _, err := api.Send(tgbotapi.NewMessage(chatID, text)); err != nil {
			log.Printf("notify chat %d: %v", chatID, err)
		}
	}

	// Register commands so they appear in Telegram's command menu / autocomplete.
	cmds := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "now_playing", Description: "Show the current track"},
		tgbotapi.BotCommand{Command: "kwewe", Description: "Show the queue"},
		tgbotapi.BotCommand{Command: "clear", Description: "Clear the queue"},
		tgbotapi.BotCommand{Command: "next", Description: "Play the next track"},
		tgbotapi.BotCommand{Command: "skip", Description: "Skip the current track"},
		tgbotapi.BotCommand{Command: "help", Description: "Show help"},
	)
	if _, err := api.Request(cmds); err != nil {
		log.Printf("set commands: %v", err)
	}
	return b, nil
}

// Username returns the bot's @handle, useful for logging.
func (b *Bot) Username() string { return b.api.Self.UserName }

// Run polls Telegram for updates until the channel closes.
func (b *Bot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		b.handle(update.Message)
	}
}

func (b *Bot) isAuthorized(chatID int64) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.authorized[chatID]
}

func (b *Bot) handle(msg *tgbotapi.Message) {
	log.Printf("request: chat=%d from=%q text=%q", msg.Chat.ID, userName(msg.From), msg.Text)

	if b.passphrase != "" {
		if !b.isAuthorized(msg.Chat.ID) {
			if strings.TrimSpace(msg.Text) == b.passphrase {
				b.mu.Lock()
				b.authorized[msg.Chat.ID] = true
				b.mu.Unlock()
				log.Printf("auth: chat %d authorized", msg.Chat.ID)
				b.reply(msg.Chat.ID, "✅ Access granted! "+helpText)
			} else {
				b.reply(msg.Chat.ID, "🔒 This bot is private. Send the passphrase to get access.")
			}
			return
		}
	} else {
		b.reply(msg.Chat.ID, "🔒 This bot is private. Contact the owner for access.")
		return
	}

	if msg.IsCommand() {
		b.handleCommand(msg)
		return
	}

	if links := ytURL.FindAllString(msg.Text, -1); len(links) > 0 {
		for _, link := range links {
			b.enqueue(msg, link)
		}
		return
	}

	b.reply(msg.Chat.ID, "Send me a YouTube / YouTube Music link to queue it, or /help for commands.")
}

func (b *Bot) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start", "help":
		b.reply(msg.Chat.ID, helpText)
	case "queue", "list", "kwewe":
		b.reply(msg.Chat.ID, b.renderQueue())
	case "now", "playing", "now_playing":
		if cur := b.player.Current(); cur != nil {
			b.reply(msg.Chat.ID, "▶️ Now playing: "+cur.Label())
		} else {
			b.reply(msg.Chat.ID, "Nothing is playing right now.")
		}
	case "skip", "next":
		if title, ok := b.player.Skip(); ok {
			b.reply(msg.Chat.ID, "⏭ Skipped: "+title)
		} else {
			b.reply(msg.Chat.ID, "Nothing is playing — nothing to skip to.")
		}
	case "clear":
		n := b.q.Clear()
		log.Printf("queue: cleared %d track(s) (by %s)", n, userName(msg.From))
		b.reply(msg.Chat.ID, fmt.Sprintf("🗑 Cleared %d track(s) from the queue.", n))
	default:
		b.reply(msg.Chat.ID, "Unknown command. Try /help.")
	}
}

func (b *Bot) enqueue(msg *tgbotapi.Message, link string) {
	addedBy := userName(msg.From)

	meta := ytinfo.Lookup(link) // best-effort; empty title falls back to URL
	t := queue.Track{URL: link, Title: meta.Title, Artist: meta.Artist, AddedBy: addedBy, ChatID: msg.Chat.ID}
	b.q.Add(t)

	ahead := b.q.Len() - 1
	log.Printf("queue: added %q (by %s) — %d ahead, %d total", t.Label(), addedBy, ahead, b.q.Len())

	// The player announces "Now playing" when a track actually starts, so we
	// just confirm the enqueue here. Len() is the pending count after adding.
	b.reply(msg.Chat.ID, fmt.Sprintf("➕ Queued (%d ahead): %s", ahead, t.Label()))
}

// userName returns a display name for a Telegram user, preferring their full
// name and falling back to their @username.
func userName(u *tgbotapi.User) string {
	if u == nil {
		return "unknown"
	}
	if name := strings.TrimSpace(u.FirstName + " " + u.LastName); name != "" {
		return name
	}
	if u.UserName != "" {
		return u.UserName
	}
	return "unknown"
}

func (b *Bot) renderQueue() string {
	items := b.q.List()
	cur := b.player.Current()

	var sb strings.Builder
	if cur != nil {
		fmt.Fprintf(&sb, "▶️ Now playing: %s\n", cur.Label())
	}
	if len(items) == 0 {
		if cur == nil {
			return "Queue is empty."
		}
		sb.WriteString("Queue is empty.")
		return sb.String()
	}
	sb.WriteString("📋 Up next:\n")
	for i, t := range items {
		fmt.Fprintf(&sb, "%d. %s — added by %s\n", i+1, t.Label(), t.AddedBy)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func (b *Bot) reply(chatID int64, text string) {
	if _, err := b.api.Send(tgbotapi.NewMessage(chatID, text)); err != nil {
		log.Printf("reply chat %d: %v", chatID, err)
	}
}

const helpText = `🎵 Music Queue Bot

Send me a YouTube or YouTube Music link and I'll add it to the queue and play tracks in order.

Commands:
/now_playing — show the current track
/kwewe       — show the queue
/next        — play the next track
/clear       — empty the queue
/skip        — skip the current track
/help        — show this message`
