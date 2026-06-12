// Package player consumes tracks from a queue and plays them in order via mpv.
package player

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"music-kwewe/internal/queue"
)

// Player plays one track at a time, in order, from a queue.
type Player struct {
	q *queue.Queue

	mu       sync.Mutex
	current  *queue.Track
	cancel   context.CancelFunc // cancels the in-flight mpv process
	ipc      net.Conn           // command channel to the playing mpv, nil until connected
	paused   bool               // whether mpv is paused, fed by watchProgress
	elapsed  float64            // seconds into the current track, fed by watchProgress
	duration float64            // total seconds of the current track, 0 when unknown

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

// Run blocks forever, playing queued tracks one after another.
func (p *Player) Run(ctx context.Context) {
	for {
		t := p.q.Next()
		p.play(ctx, t)
	}
}

func (p *Player) play(ctx context.Context, t queue.Track) {
	playCtx, cancel := context.WithCancel(ctx)

	p.mu.Lock()
	p.current = &t
	p.cancel = cancel
	p.paused = false
	p.elapsed, p.duration = 0, 0
	p.mu.Unlock()

	log.Printf("now playing: %q (%s)", t.Label(), t.URL)
	p.notify(t.ChatID, "▶️ Now playing: "+t.Label())
	if p.OnPlay != nil {
		p.OnPlay(t)
	}

	// mpv plays YouTube/YouTube Music URLs directly (it shells out to yt-dlp).
	// --no-video: audio only. --no-terminal: don't grab our stdin.
	// --input-ipc-server: JSON IPC socket watched for playback progress.
	sock := filepath.Join(os.TempDir(), fmt.Sprintf("kwewe-mpv-%d.sock", os.Getpid()))
	_ = os.Remove(sock) // clear a stale socket from a previous run
	cmd := exec.CommandContext(playCtx, "mpv", "--no-video", "--no-terminal", "--input-ipc-server="+sock, t.URL)
	err := cmd.Start()
	if err == nil {
		go p.watchProgress(playCtx, sock)
		err = cmd.Wait()
	}

	p.mu.Lock()
	skipped := p.current == nil // Skip() clears current before killing
	p.current = nil
	p.cancel = nil
	p.mu.Unlock()
	cancel()
	_ = os.Remove(sock)

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

// Progress returns seconds elapsed and total for the current track. ok is
// false when nothing is playing; duration may be 0 when mpv doesn't know it
// yet (or ever — e.g. live streams).
func (p *Player) Progress() (elapsed, duration float64, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current == nil {
		return 0, 0, false
	}
	return p.elapsed, p.duration, true
}

// Paused reports whether the current track is paused. False when idle.
func (p *Player) Paused() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current != nil && p.paused
}

// TogglePause flips mpv between paused and playing. paused reports the new
// state. ok is false when idle or before the IPC connection is up.
func (p *Player) TogglePause() (paused bool, ok bool) {
	p.mu.Lock()
	conn := p.ipc
	if p.current == nil || conn == nil {
		p.mu.Unlock()
		return false, false
	}
	paused = !p.paused
	p.paused = paused // optimistic; the observed pause event confirms it
	p.mu.Unlock()

	if _, err := fmt.Fprintf(conn, `{"command":["set_property","pause",%t]}`+"\n", paused); err != nil {
		return false, false
	}
	return paused, true
}

// watchProgress connects to mpv's IPC socket and mirrors the time-pos,
// duration, and pause properties into the Player until playback ends. The
// connection doubles as the command channel for TogglePause.
func (p *Player) watchProgress(ctx context.Context, sock string) {
	conn := dialIPC(ctx, sock)
	if conn == nil {
		return
	}
	defer conn.Close()
	go func() { // unblock the scanner when the track ends or is skipped
		<-ctx.Done()
		conn.Close()
	}()

	p.mu.Lock()
	p.ipc = conn
	p.mu.Unlock()
	defer func() {
		p.mu.Lock()
		if p.ipc == conn { // don't clobber the next track's connection
			p.ipc = nil
		}
		p.mu.Unlock()
	}()

	fmt.Fprintln(conn, `{"command":["observe_property",1,"time-pos"]}`)
	fmt.Fprintln(conn, `{"command":["observe_property",2,"duration"]}`)
	fmt.Fprintln(conn, `{"command":["observe_property",3,"pause"]}`)

	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		name, val, ok := parseProgressEvent(sc.Bytes())
		if !ok {
			continue
		}
		p.mu.Lock()
		switch name {
		case "time-pos":
			p.elapsed = val.(float64)
		case "duration":
			p.duration = val.(float64)
		case "pause":
			p.paused = val.(bool)
		}
		p.mu.Unlock()
	}
}

// dialIPC dials mpv's unix socket, retrying while mpv starts up (resolving a
// YouTube URL via yt-dlp can take a few seconds). Returns nil if the socket
// never appears or ctx is cancelled first.
func dialIPC(ctx context.Context, sock string) net.Conn {
	for i := 0; i < 40; i++ {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			return conn
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(250 * time.Millisecond):
		}
	}
	return nil
}

// parseProgressEvent extracts a watched property value from one line of mpv
// IPC output: float64 for time-pos/duration, bool for pause. ok is false for
// everything else: command replies, other events, and null data (mpv sends
// null when a property is unavailable).
func parseProgressEvent(line []byte) (name string, val any, ok bool) {
	var ev struct {
		Event string          `json:"event"`
		Name  string          `json:"name"`
		Data  json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(line, &ev); err != nil {
		return "", nil, false
	}
	if ev.Event != "property-change" || len(ev.Data) == 0 || string(ev.Data) == "null" {
		return "", nil, false
	}
	switch ev.Name {
	case "time-pos", "duration":
		var f float64
		if json.Unmarshal(ev.Data, &f) != nil {
			return "", nil, false
		}
		return ev.Name, f, true
	case "pause":
		var b bool
		if json.Unmarshal(ev.Data, &b) != nil {
			return "", nil, false
		}
		return ev.Name, b, true
	}
	return "", nil, false
}

// FormatClock renders seconds as m:ss, or h:mm:ss from one hour up.
func FormatClock(seconds float64) string {
	s := int(seconds)
	if s < 0 {
		s = 0
	}
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, (s%3600)/60, s%60)
	}
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}

func (p *Player) notify(chatID int64, text string) {
	if p.Notify != nil {
		p.Notify(chatID, text)
	}
}
