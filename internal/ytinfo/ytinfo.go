// Package ytinfo resolves human-readable metadata (title, artist) of a URL.
package ytinfo

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// Meta is the resolved metadata for a URL. Any field that can't be resolved
// comes back empty so callers can fall back gracefully.
type Meta struct {
	Title  string // video title
	Artist string // uploading channel / artist
}

// Lookup resolves the title and channel of url using yt-dlp. It is best-effort:
// on any error it returns a zero Meta so callers can fall back to the URL.
func Lookup(url string) Meta {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// One yt-dlp call prints the title then the channel, each on its own line.
	cmd := exec.CommandContext(ctx, "yt-dlp", "--no-playlist", "--skip-download",
		"--print", "title", "--print", "channel", url)
	out, err := cmd.Output()
	if err != nil {
		return Meta{}
	}

	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	var m Meta
	if len(lines) > 0 {
		m.Title = strings.TrimSpace(lines[0])
	}
	if len(lines) > 1 {
		// yt-dlp prints "NA" when a field is unavailable; treat that as empty.
		if artist := strings.TrimSpace(lines[1]); artist != "NA" {
			m.Artist = artist
		}
	}
	return m
}
