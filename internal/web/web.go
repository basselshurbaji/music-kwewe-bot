// Package web serves a small read-only dashboard showing the now-playing
// track and the pending queue.
package web

import (
	"encoding/json"
	"net/http"

	qrc "github.com/skip2/go-qrcode"

	"music-kwewe/internal/player"
	"music-kwewe/internal/queue"
	"music-kwewe/internal/stats"
)

// Server exposes the queue, player, and session stats over HTTP.
type Server struct {
	q          *queue.Queue
	p          *player.Player
	st         *stats.Stats
	botLink    string
	passphrase string
}

// New returns a dashboard server bound to the given queue, player, and stats.
// botLink and passphrase are surfaced via /api/invite and /qr.
func New(q *queue.Queue, p *player.Player, st *stats.Stats, botLink, passphrase string) *Server {
	return &Server{q: q, p: p, st: st, botLink: botLink, passphrase: passphrase}
}

// Page returns the dashboard HTML. Exposed so tooling (e.g. a seeded preview)
// can reuse the exact page the server serves.
func Page() string { return indexHTML }

// Handler returns the HTTP routes.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.index)
	mux.HandleFunc("/api/state", s.state)
	mux.HandleFunc("/api/invite", s.invite)
	mux.HandleFunc("/qr", s.qr)
	return mux
}

type trackView struct {
	Title   string `json:"title"`
	AddedBy string `json:"added_by"`
	URL     string `json:"url"`
	// Elapsed/Duration (whole seconds) and Paused are set only on the
	// now-playing track.
	Elapsed  int  `json:"elapsed,omitempty"`
	Duration int  `json:"duration,omitempty"`
	Paused   bool `json:"paused,omitempty"`
}

type stateView struct {
	NowPlaying   *trackView   `json:"now_playing"`
	Queue        []trackView  `json:"queue"`
	Contributors []stats.Stat `json:"contributors"`
	Artists      []stats.Stat `json:"artists"`
	Played       int          `json:"played"`
}

func (s *Server) state(w http.ResponseWriter, _ *http.Request) {
	view := stateView{Queue: []trackView{}, Contributors: []stats.Stat{}, Artists: []stats.Stat{}}
	if cur := s.p.Current(); cur != nil {
		view.NowPlaying = &trackView{Title: cur.Label(), AddedBy: cur.AddedBy, URL: cur.URL}
		if elapsed, duration, ok := s.p.Progress(); ok {
			view.NowPlaying.Elapsed = int(elapsed)
			view.NowPlaying.Duration = int(duration)
		}
		view.NowPlaying.Paused = s.p.Paused()
	}
	for _, t := range s.q.List() {
		view.Queue = append(view.Queue, trackView{Title: t.Label(), AddedBy: t.AddedBy, URL: t.URL})
	}
	if s.st != nil {
		view.Contributors = s.st.TopContributors(5)
		view.Artists = s.st.TopArtists(5)
		view.Played = s.st.Played()
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(view)
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(indexHTML))
}

type inviteView struct {
	BotLink    string `json:"bot_link"`
	Passphrase string `json:"passphrase"`
}

func (s *Server) invite(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(inviteView{BotLink: s.botLink, Passphrase: s.passphrase})
}

func (s *Server) qr(w http.ResponseWriter, r *http.Request) {
	if s.botLink == "" {
		http.NotFound(w, r)
		return
	}
	png, err := qrc.Encode(s.botLink, qrc.Medium, 256)
	if err != nil {
		http.Error(w, "could not generate QR", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write(png)
}
