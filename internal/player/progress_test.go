package player

import "testing"

func TestParseProgressEvent(t *testing.T) {
	tests := []struct {
		line string
		name string
		val  any
		ok   bool
	}{
		{`{"event":"property-change","id":1,"name":"time-pos","data":42.5}`, "time-pos", 42.5, true},
		{`{"event":"property-change","id":2,"name":"duration","data":225}`, "duration", 225.0, true},
		{`{"event":"property-change","id":3,"name":"pause","data":true}`, "pause", true, true},
		{`{"event":"property-change","id":3,"name":"pause","data":false}`, "pause", false, true},
		{`{"event":"property-change","id":3,"name":"pause","data":"yes"}`, "", nil, false},
		{`{"event":"property-change","id":1,"name":"time-pos","data":null}`, "", nil, false},
		{`{"event":"property-change","id":1,"name":"time-pos"}`, "", nil, false},
		{`{"event":"property-change","id":4,"name":"volume","data":100}`, "", nil, false},
		{`{"request_id":0,"error":"success"}`, "", nil, false},
		{`{"event":"end-file"}`, "", nil, false},
		{`not json`, "", nil, false},
		{``, "", nil, false},
	}
	for _, tt := range tests {
		name, val, ok := parseProgressEvent([]byte(tt.line))
		if name != tt.name || val != tt.val || ok != tt.ok {
			t.Errorf("parseProgressEvent(%q) = (%q, %v, %v), want (%q, %v, %v)",
				tt.line, name, val, ok, tt.name, tt.val, tt.ok)
		}
	}
}

func TestFormatClock(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0:00"},
		{7.9, "0:07"},
		{65, "1:05"},
		{600, "10:00"},
		{3599, "59:59"},
		{3600, "1:00:00"},
		{3661.5, "1:01:01"},
		{-3, "0:00"},
	}
	for _, tt := range tests {
		if got := FormatClock(tt.in); got != tt.want {
			t.Errorf("FormatClock(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestProgressIdle(t *testing.T) {
	p := New(nil)
	if _, _, ok := p.Progress(); ok {
		t.Error("Progress() ok = true on an idle player, want false")
	}
}

func TestTogglePauseIdle(t *testing.T) {
	p := New(nil)
	if _, ok := p.TogglePause(); ok {
		t.Error("TogglePause() ok = true on an idle player, want false")
	}
	if p.Paused() {
		t.Error("Paused() = true on an idle player, want false")
	}
}
