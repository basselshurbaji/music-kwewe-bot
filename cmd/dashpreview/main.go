// Command dashpreview serves the dashboard with seeded sample data (including
// a now-playing track) for visual checks and screenshots. Not part of the
// running service.
package main

import (
	"log"
	"net/http"

	"music-queue/internal/web"
)

const sampleState = `{
  "now_playing": { "title": "Led Zeppelin - Stairway to Heaven", "added_by": "Ada", "url": "https://www.youtube.com/watch?v=qZTh8DK7jdk" },
  "queue": [
    { "title": "Queen - Bohemian Rhapsody", "added_by": "Linus", "url": "https://www.youtube.com/watch?v=fJ9rUzIMcZQ" },
    { "title": "Pink Floyd - Comfortably Numb", "added_by": "Grace", "url": "https://www.youtube.com/watch?v=_FrOQC-zEog" },
    { "title": "The Rolling Stones - Paint It Black", "added_by": "Dennis", "url": "https://www.youtube.com/watch?v=O4irXQhgMqg" },
    { "title": "AC/DC - Back in Black", "added_by": "Alan", "url": "https://www.youtube.com/watch?v=pAgnJDJN4VA" },
    { "title": "Jimi Hendrix - All Along the Watchtower", "added_by": "Margaret", "url": "https://www.youtube.com/watch?v=TLV4_xaYynY" }
  ]
}`

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(web.Page()))
	})
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleState))
	})

	log.Println("preview on :7171")
	log.Fatal(http.ListenAndServe(":7171", mux))
}
