// Package stats tracks per-session play counts: who queued the most tracks
// (top contributors) and which YouTube channels/artists got the most airtime.
// Everything is in-memory and resets when the process restarts.
package stats

import (
	"sort"
	"sync"
)

// Stat is a name paired with how many plays it accounts for.
type Stat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Stats holds session-wide play tallies, keyed by contributor and by artist.
type Stats struct {
	mu           sync.Mutex
	contributors map[string]int
	artists      map[string]int
	played       int
}

// New returns an empty, ready-to-use Stats.
func New() *Stats {
	return &Stats{
		contributors: make(map[string]int),
		artists:      make(map[string]int),
	}
}

// Record counts one played track. Either name may be empty (e.g. an unknown
// artist), in which case that dimension is simply not tallied.
func (s *Stats) Record(addedBy, artist string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.played++
	if addedBy != "" {
		s.contributors[addedBy]++
	}
	if artist != "" {
		s.artists[artist]++
	}
}

// TopContributors returns the n biggest queuers, highest count first. n <= 0
// returns all of them.
func (s *Stats) TopContributors(n int) []Stat {
	s.mu.Lock()
	defer s.mu.Unlock()
	return topN(s.contributors, n)
}

// TopArtists returns the n most-played channels/artists, highest count first.
// n <= 0 returns all of them.
func (s *Stats) TopArtists(n int) []Stat {
	s.mu.Lock()
	defer s.mu.Unlock()
	return topN(s.artists, n)
}

// Played returns the total number of tracks played this session.
func (s *Stats) Played() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.played
}

// topN sorts a count map by count (desc), then name (asc) for stable ties, and
// trims to n entries. Caller must hold the lock.
func topN(m map[string]int, n int) []Stat {
	out := make([]Stat, 0, len(m))
	for name, count := range m {
		out = append(out, Stat{Name: name, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})
	if n > 0 && len(out) > n {
		out = out[:n]
	}
	return out
}
