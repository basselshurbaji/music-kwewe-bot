package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"music-kwewe/internal/player"
	"music-kwewe/internal/queue"
	"music-kwewe/internal/stats"
)

func newServer() *Server {
	q := queue.New()
	q.Add(queue.Track{URL: "https://www.youtube.com/watch?v=aaaaaaaaaaa", Title: "First", AddedBy: "Bassel"})
	q.Add(queue.Track{URL: "https://www.youtube.com/watch?v=bbbbbbbbbbb", Title: "Second", AddedBy: "Sam"})
	return New(q, player.New(q), stats.New())
}

func TestStateJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	newServer().Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type = %q", ct)
	}

	var s stateView
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if s.NowPlaying != nil {
		t.Errorf("now_playing = %+v, want nil (idle)", s.NowPlaying)
	}
	if len(s.Queue) != 2 {
		t.Fatalf("queue len = %d, want 2", len(s.Queue))
	}
	if s.Queue[0].Title != "First" || s.Queue[0].AddedBy != "Bassel" {
		t.Errorf("queue[0] = %+v", s.Queue[0])
	}
}

func TestIndexHTML(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	newServer().Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "music-kwewe-bot") {
		t.Error("html missing title")
	}
	if !strings.Contains(body, "/api/state") {
		t.Error("html missing polling endpoint")
	}
}

func TestNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	newServer().Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
