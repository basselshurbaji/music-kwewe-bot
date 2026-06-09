// Package player consumes tracks from a queue and plays them in order via mpv.
package player

import (
	"context"
	"log"
	"os/exec"
	"sync"

	"music-queue/internal/queue"
)

// Player plays one track at a time, in order, from a queue.
type Player struct {
	q *queue.Queue

	mu      sync.Mutex
	current *queue.Track
	cancel  context.CancelFunc // cancels the in-flight mpv process

	// Notify, if set, is called on playback state changes so the bot can
	// message the relevant chat. Safe to leave nil.
	Notify func(chatID int64, text string)

	// OnPlay, if set, is called once when a track starts playing. Used to feed
	// session stats. Safe to leave nil.
	OnPlay func(t queue.Track)
}

// New returns a Player bound to the given queue.
func New(q *queue.Queue) *Player {
	return &Player{q: q}
}

// Run blocks forever, playing queued tracks one after another. Call it in a
// goroutine. It stops when ctx is cancelled.
func (p *Player) Run(ctx context.Context) {
	for {
		// queue.Next blocks; bail out if we're shutting down.
		select {
		case <-ctx.Done():
			return
		default:
		}
		t := p.q.Next()
		p.play(ctx, t)
	}
}

func (p *Player) play(ctx context.Context, t queue.Track) {
	playCtx, cancel := context.WithCancel(ctx)

	p.mu.Lock()
	p.current = &t
	p.cancel = cancel
	p.mu.Unlock()

	log.Printf("now playing: %q (%s)", t.Label(), t.URL)
	p.notify(t.ChatID, "▶️ Now playing: "+t.Label())
	if p.OnPlay != nil {
		p.OnPlay(t)
	}

	// mpv plays YouTube/YouTube Music URLs directly (it shells out to yt-dlp).
	// --no-video: audio only. --no-terminal: don't grab our stdin.
	cmd := exec.CommandContext(playCtx, "mpv", "--no-video", "--no-terminal", t.URL)
	err := cmd.Run()

	p.mu.Lock()
	skipped := p.current == nil // Skip() clears current before killing
	p.current = nil
	p.cancel = nil
	p.mu.Unlock()
	cancel()

	switch {
	case skipped:
		log.Printf("skipped: %q", t.Label())
		// Already messaged by Skip(); nothing to do.
	case err != nil && playCtx.Err() == nil:
		log.Printf("playback error for %s: %v", t.URL, err)
		p.notify(t.ChatID, "⚠️ Failed to play: "+t.Label())
	default:
		log.Printf("finished: %q", t.Label())
		p.notify(t.ChatID, "✅ Finished: "+t.Label())
	}
}

// Skip stops the currently playing track (if any) and returns its title.
// The Run loop then advances to the next queued track.
func (p *Player) Skip() (title string, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current == nil || p.cancel == nil {
		return "", false
	}
	title = p.current.Label()
	p.current = nil // mark as intentional skip so play() stays quiet
	p.cancel()
	p.cancel = nil
	return title, true
}

// Current returns the track playing right now, or nil if idle.
func (p *Player) Current() *queue.Track {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current
}

func (p *Player) notify(chatID int64, text string) {
	if p.Notify != nil {
		p.Notify(chatID, text)
	}
}
