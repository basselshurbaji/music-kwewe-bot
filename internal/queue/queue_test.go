package queue

import (
	"testing"
	"time"
)

func TestFIFOOrder(t *testing.T) {
	q := New()
	q.Add(Track{URL: "a"})
	q.Add(Track{URL: "b"})
	q.Add(Track{URL: "c"})

	for _, want := range []string{"a", "b", "c"} {
		if got := q.Next().URL; got != want {
			t.Fatalf("Next() = %q, want %q", got, want)
		}
	}
	if q.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", q.Len())
	}
}

func TestNextBlocksUntilAdd(t *testing.T) {
	q := New()
	done := make(chan string, 1)
	go func() { done <- q.Next().URL }()

	select {
	case <-done:
		t.Fatal("Next returned before any track was added")
	case <-time.After(50 * time.Millisecond):
	}

	q.Add(Track{URL: "x"})
	select {
	case got := <-done:
		if got != "x" {
			t.Fatalf("Next() = %q, want %q", got, "x")
		}
	case <-time.After(time.Second):
		t.Fatal("Next did not return after Add")
	}
}

func TestClear(t *testing.T) {
	q := New()
	q.Add(Track{URL: "a"})
	q.Add(Track{URL: "b"})
	if n := q.Clear(); n != 2 {
		t.Fatalf("Clear() = %d, want 2", n)
	}
	if q.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", q.Len())
	}
}

func TestLabelFallback(t *testing.T) {
	if got := (Track{URL: "u"}).Label(); got != "u" {
		t.Fatalf("Label() = %q, want url fallback", got)
	}
	if got := (Track{URL: "u", Title: "T"}).Label(); got != "T" {
		t.Fatalf("Label() = %q, want title", got)
	}
}
