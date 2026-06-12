package bot

import (
	"testing"

	"music-kwewe/internal/queue"
)

func TestFindDuplicate(t *testing.T) {
	watch := queue.Track{URL: "https://www.youtube.com/watch?v=abc123", ID: "abc123", Title: "Song A"}
	short := queue.Track{URL: "https://youtu.be/abc123?si=xyz", ID: "abc123", Title: "Song A"}
	other := queue.Track{URL: "https://www.youtube.com/watch?v=def456", ID: "def456", Title: "Song B"}
	noID := queue.Track{URL: "https://www.youtube.com/watch?v=ghi789", Title: "Song C"}

	tests := []struct {
		name        string
		cur         *queue.Track
		items       []queue.Track
		t           queue.Track
		wantDup     bool
		wantPos     int
		wantPlaying bool
	}{
		{name: "empty queue, nothing playing", t: watch},
		{name: "same ID via different URL form", items: []queue.Track{watch}, t: short, wantDup: true},
		{name: "matches now playing", cur: &watch, t: short, wantDup: true, wantPlaying: true},
		{name: "different IDs", cur: &other, items: []queue.Track{other}, t: watch},
		{name: "position of later match", items: []queue.Track{other, watch}, t: short, wantDup: true, wantPos: 1},
		{name: "no ID, exact URL match", items: []queue.Track{noID}, t: noID, wantDup: true},
		{name: "no ID, different URLs", items: []queue.Track{noID}, t: queue.Track{URL: "https://youtu.be/ghi789"}},
		{name: "one side missing ID, same URL", items: []queue.Track{noID}, t: queue.Track{URL: noID.URL, ID: "ghi789"}, wantDup: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dup, pos, playing := findDuplicate(tc.cur, tc.items, tc.t)
			if (dup != nil) != tc.wantDup {
				t.Fatalf("dup = %v, want match=%v", dup, tc.wantDup)
			}
			if pos != tc.wantPos {
				t.Errorf("pos = %d, want %d", pos, tc.wantPos)
			}
			if playing != tc.wantPlaying {
				t.Errorf("playing = %v, want %v", playing, tc.wantPlaying)
			}
		})
	}
}
