# music-kwewe-bot

A Go service that runs a Telegram bot. Send it YouTube / YouTube Music links;
it queues them and plays the audio in order on the machine running the service.

## How it works

```
Telegram → bot (links + commands) → FIFO queue → player goroutine → mpv (audio)
```

- **bot** parses messages: any YouTube/`youtu.be`/`music.youtube.com` URL is enqueued; slash commands control playback.
- **queue** is an in-memory, thread-safe FIFO. `Next()` blocks until a track is available.
- **player** pulls one track at a time and plays it with `mpv --no-video`. mpv shells out to `yt-dlp` to resolve the stream, so YouTube links play directly. Playback state changes are pushed back to the requesting Telegram chat.

The queue is in-memory only — it does not survive a restart.

## Requirements

- Go 1.26+
- [`mpv`](https://mpv.io/) — audio playback
- [`yt-dlp`](https://github.com/yt-dlp/yt-dlp) — used by mpv to resolve YouTube streams, and by the bot to fetch track titles

```sh
brew install mpv yt-dlp
```

Audio plays through the **default output device of the host running the service**.

## Setup

1. Install dependencies and create your local `.env`:

   ```sh
   make setup
   ```

   This runs `scripts/install.sh` (Homebrew: `mpv`, `yt-dlp`) and copies
   `.env.example` to `.env`.

2. Create a bot with [@BotFather](https://t.me/BotFather), copy the token, and
   put it in `.env`:

   ```
   TELEGRAM_BOT_TOKEN=123456:abc...
   ```

3. Run it:

   ```sh
   make run
   ```

`.env` is loaded automatically at startup and is gitignored. A real environment
variable, if set, takes precedence over the `.env` value.

### Make targets

| Command | Does |
|---|---|
| `make setup` | Install Homebrew deps + create `.env` from template |
| `make run`   | Run the service (`go run .`) |
| `make build` | Build the `music-queue` binary |
| `make test`  | Run tests |
| `make clean` | Remove the built binary |

## Usage

Message the bot:

- Paste a YouTube / YouTube Music link → it's added to the queue (multiple links in one message all get queued).
- `/queue` — show what's playing and what's up next
- `/now` — show the current track
- `/skip` — skip the current track
- `/clear` — empty the queue
- `/help` — list commands

## Tests

```sh
go test ./...
```
