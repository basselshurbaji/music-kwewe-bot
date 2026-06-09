// Package ytinfo resolves the human-readable title of a YouTube URL.
package ytinfo

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// Title returns the video title for url using yt-dlp. It is best-effort:
// on any error it returns an empty string so callers can fall back to the URL.
func Title(url string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp", "--no-playlist", "--skip-download", "--print", "title", url)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
