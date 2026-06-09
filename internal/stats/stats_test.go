package stats

import "testing"

func TestRecordAndRank(t *testing.T) {
	s := New()
	s.Record("Ada", "Queen")
	s.Record("Ada", "Led Zeppelin")
	s.Record("Linus", "Queen")
	s.Record("Ada", "Queen")

	if got := s.Played(); got != 4 {
		t.Fatalf("Played() = %d, want 4", got)
	}

	djs := s.TopContributors(0)
	if len(djs) != 2 || djs[0].Name != "Ada" || djs[0].Count != 3 {
		t.Fatalf("TopContributors = %+v, want Ada=3 first", djs)
	}

	artists := s.TopArtists(0)
	if artists[0].Name != "Queen" || artists[0].Count != 3 {
		t.Fatalf("TopArtists = %+v, want Queen=3 first", artists)
	}
}

func TestEmptyNamesIgnored(t *testing.T) {
	s := New()
	s.Record("", "")     // unknown DJ and artist
	s.Record("Ada", "")  // known DJ, unknown artist

	if got := s.Played(); got != 2 {
		t.Fatalf("Played() = %d, want 2", got)
	}
	if djs := s.TopContributors(0); len(djs) != 1 || djs[0].Name != "Ada" {
		t.Fatalf("TopContributors = %+v, want only Ada", djs)
	}
	if artists := s.TopArtists(0); len(artists) != 0 {
		t.Fatalf("TopArtists = %+v, want none", artists)
	}
}

func TestTopNLimit(t *testing.T) {
	s := New()
	for _, name := range []string{"a", "b", "c", "d"} {
		s.Record(name, "")
	}
	if got := s.TopContributors(2); len(got) != 2 {
		t.Fatalf("TopContributors(2) len = %d, want 2", len(got))
	}
}
