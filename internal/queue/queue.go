// Package queue implements a thread-safe FIFO queue of music tracks.
package queue

import "sync"

// Track is a single queued item.
type Track struct {
	URL     string
	Title   string
	AddedBy string // Telegram display name of whoever queued it
	ChatID  int64  // chat the track was requested from (for notifications)
}

// Label returns a human-friendly name for the track, falling back to the URL.
func (t Track) Label() string {
	if t.Title != "" {
		return t.Title
	}
	return t.URL
}

// Queue is a thread-safe FIFO queue. Next blocks until a track is available.
type Queue struct {
	mu    sync.Mutex
	cond  *sync.Cond
	items []Track
}

// New returns an empty queue.
func New() *Queue {
	q := &Queue{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Add appends a track and wakes any goroutine blocked in Next.
func (q *Queue) Add(t Track) {
	q.mu.Lock()
	q.items = append(q.items, t)
	q.mu.Unlock()
	q.cond.Signal()
}

// Next removes and returns the front track, blocking until one exists.
func (q *Queue) Next() Track {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == 0 {
		q.cond.Wait()
	}
	t := q.items[0]
	q.items = q.items[1:]
	return t
}

// List returns a snapshot copy of the pending tracks.
func (q *Queue) List() []Track {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]Track, len(q.items))
	copy(out, q.items)
	return out
}

// Clear empties the queue and returns how many tracks were removed.
func (q *Queue) Clear() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	n := len(q.items)
	q.items = nil
	return n
}

// Len returns the number of pending tracks.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}
