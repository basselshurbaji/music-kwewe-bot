// Package web serves a small read-only dashboard showing the now-playing
// track and the pending queue.
package web

import (
	"encoding/json"
	"net/http"

	"music-kwewe/internal/player"
	"music-kwewe/internal/queue"
	"music-kwewe/internal/stats"
)

// Server exposes the queue, player, and session stats over HTTP.
type Server struct {
	q  *queue.Queue
	p  *player.Player
	st *stats.Stats
}

// New returns a dashboard server bound to the given queue, player, and stats.
func New(q *queue.Queue, p *player.Player, st *stats.Stats) *Server {
	return &Server{q: q, p: p, st: st}
}

// Page returns the dashboard HTML. Exposed so tooling (e.g. a seeded preview)
// can reuse the exact page the server serves.
func Page() string { return indexHTML }

// Handler returns the HTTP routes: "/" (dashboard) and "/api/state" (JSON).
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.index)
	mux.HandleFunc("/api/state", s.state)
	return mux
}

type trackView struct {
	Title   string `json:"title"`
	AddedBy string `json:"added_by"`
	URL     string `json:"url"`
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
	}
	for _, t := range s.q.List() {
		view.Queue = append(view.Queue, trackView{Title: t.Label(), AddedBy: t.AddedBy, URL: t.URL})
	}
	if s.st != nil {
		view.Contributors = s.st.TopContributors(8)
		view.Artists = s.st.TopArtists(8)
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
	_, _ = w.Write([]byte(indexHTML))
}
