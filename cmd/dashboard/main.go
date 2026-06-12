// Command dashboard is a terminal TUI that mirrors the web dashboard.
// It polls the bot's /api/state endpoint and re-renders every 2 seconds.
// Run the bot first (make run), then open this in a second terminal (make dashboard).
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"music-kwewe/internal/player"
)

// ── palette ──────────────────────────────────────────────────────────────────

var (
	colBg     = lipgloss.Color("#0a0e14")
	colFg     = lipgloss.Color("#c6ffd6")
	colDim    = lipgloss.Color("#4b6e57")
	colAccent = lipgloss.Color("#00ff9c")
	colAmber  = lipgloss.Color("#ffcb6b")
	colRed    = lipgloss.Color("#ff6b6b")
	colBorder = lipgloss.Color("#1c2b22")
)

// ── styles ────────────────────────────────────────────────────────────────────

var (
	stDim    = lipgloss.NewStyle().Foreground(colDim)
	stAccent = lipgloss.NewStyle().Foreground(colAccent)
	stAmber  = lipgloss.NewStyle().Foreground(colAmber)
	stFg     = lipgloss.NewStyle().Foreground(colFg)
	stRed    = lipgloss.NewStyle().Foreground(colRed)
	stBox = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colBorder)

	stScreen = lipgloss.NewStyle().
			PaddingLeft(2).PaddingRight(2).
			PaddingTop(1).PaddingBottom(1)

	stSection = lipgloss.NewStyle().MarginTop(1)

	stHeading = lipgloss.NewStyle().
			Foreground(colDim).
			Bold(true).
			MarginBottom(1)

	stNowBar = lipgloss.NewStyle().
			BorderStyle(lipgloss.Border{Left: "│"}).
			BorderForeground(colAccent).
			BorderLeft(true).
			PaddingLeft(1)
)

// ── API types ─────────────────────────────────────────────────────────────────

type trackView struct {
	Title    string `json:"title"`
	AddedBy  string `json:"added_by"`
	URL      string `json:"url"`
	Elapsed  int    `json:"elapsed"`
	Duration int    `json:"duration"`
	Paused   bool   `json:"paused"`
}

type inviteView struct {
	BotLink    string `json:"bot_link"`
	Passphrase string `json:"passphrase"`
}

type statView struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type stateView struct {
	NowPlaying   *trackView `json:"now_playing"`
	Queue        []trackView `json:"queue"`
	Contributors []statView  `json:"contributors"`
	Artists      []statView  `json:"artists"`
	Played       int         `json:"played"`
}

// ── messages ──────────────────────────────────────────────────────────────────

type stateMsg stateView
type inviteMsg inviteView
type errMsg struct{ err error }
type tickMsg time.Time

// ── model ─────────────────────────────────────────────────────────────────────

type model struct {
	baseURL string
	state   stateView
	invite  inviteView
	err     error
	updated time.Time
	width   int
	height  int
}

func newModel() model {
	addr := os.Getenv("DASHBOARD_ADDR")
	if addr == "" {
		addr = ":7070"
	}
	port := addr
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		port = addr[i:]
	}
	return model{baseURL: "http://localhost" + port}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchCmd(m.baseURL+"/api/state"), fetchInviteCmd(m.baseURL+"/api/invite"), tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		return m, tea.Batch(fetchCmd(m.baseURL+"/api/state"), tickCmd())
	case inviteMsg:
		m.invite = inviteView(msg)
	case stateMsg:
		m.state = stateView(msg)
		m.err = nil
		m.updated = time.Now()
	case errMsg:
		m.err = msg.err
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func fetchCmd(url string) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(url) //nolint:noctx
		if err != nil {
			return errMsg{err}
		}
		defer resp.Body.Close()
		var s stateView
		if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
			return errMsg{err}
		}
		return stateMsg(s)
	}
}

func fetchInviteCmd(url string) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(url) //nolint:noctx
		if err != nil {
			return nil
		}
		defer resp.Body.Close()
		var inv inviteView
		if err := json.NewDecoder(resp.Body).Decode(&inv); err != nil {
			return nil
		}
		return inviteMsg(inv)
	}
}

// ── view ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	innerWidth := m.width - 4 // 2 box border + 2 padding each side
	if innerWidth < 40 {
		innerWidth = 40
	}

	// Bot access (only when invite info is available)
	var accessSection string
	if m.invite.BotLink != "" || m.invite.Passphrase != "" {
		var lines []string
		if m.invite.BotLink != "" {
			lines = append(lines, stDim.Render("link: ")+stFg.Render(m.invite.BotLink))
		}
		if m.invite.Passphrase != "" {
			lines = append(lines, stDim.Render("pass: ")+stAccent.Render(m.invite.Passphrase))
		}
		accessSection = stSection.Render(stHeading.Render("⌘ BOT ACCESS") + "\n" + strings.Join(lines, "\n"))
	}

	// Now playing
	nowHeading := stHeading.Render("▶ NOW PLAYING")
	var nowBody string
	if m.err != nil {
		nowBody = stRed.Render("connection lost — is the bot running?")
	} else if m.state.NowPlaying == nil {
		nowBody = stDim.Render("nothing playing")
	} else {
		np := m.state.NowPlaying
		line := stAccent.Render(np.Title)
		if np.AddedBy != "" {
			line += "\n" + stDim.Render("added by "+np.AddedBy)
		}
		if prog := renderProgress(np.Elapsed, np.Duration, np.Paused); prog != "" {
			line += "\n" + prog
		}
		nowBody = stNowBar.Render(line)
	}
	nowSection := stSection.Render(nowHeading + "\n" + nowBody)

	// Queue
	queueHeading := stHeading.Render("# QUEUE")
	var queueBody string
	if len(m.state.Queue) == 0 {
		queueBody = stDim.Render("queue is empty")
	} else {
		var lines []string
		for i, t := range m.state.Queue {
			idx := stAmber.Render(fmt.Sprintf("%2d.", i+1))
			track := stFg.Render(t.Title)
			by := ""
			if t.AddedBy != "" {
				by = stDim.Render(" — " + t.AddedBy)
			}
			lines = append(lines, idx+" "+track+by)
		}
		queueBody = strings.Join(lines, "\n")
	}
	queueSection := stSection.Render(queueHeading + "\n" + queueBody)

	// Stats — two columns side by side
	statsHeading := stHeading.Render("~/.stats — session")
	djsHeading := stHeading.Render("★ TOP DJS")
	artistsHeading := stHeading.Render("♪ SESSION ARTISTS")

	colW := (innerWidth - 4) / 2

	djsBody := renderChart(m.state.Contributors, colW)
	artistsBody := renderChart(m.state.Artists, colW)

	djsCol := lipgloss.NewStyle().Width(colW).Render(djsHeading + "\n" + djsBody)
	artistsCol := lipgloss.NewStyle().Width(colW).Render(artistsHeading + "\n" + artistsBody)
	statsGrid := lipgloss.JoinHorizontal(lipgloss.Top, djsCol, "  ", artistsCol)
	statsSection := stSection.Render(statsHeading + "\n" + statsGrid)

	// Footer
	total := len(m.state.Queue)
	if m.state.NowPlaying != nil {
		total++
	}
	updatedStr := "--:--:--"
	if !m.updated.IsZero() {
		updatedStr = m.updated.Format("15:04:05")
	}
	footerParts := []string{
		stDim.Render(fmt.Sprintf("tracks: %d", total)),
		stDim.Render(fmt.Sprintf("played: %d", m.state.Played)),
		stDim.Render("updated: " + updatedStr),
		stDim.Render("press q to quit"),
	}
	footer := stSection.Render(
		stDim.Render(strings.Repeat("─", innerWidth)) + "\n" +
			strings.Join(footerParts, stDim.Render("  ·  ")),
	)

	// Assemble screen content
	body := nowSection + "\n" + queueSection + "\n" + statsSection + "\n" + footer
	if accessSection != "" {
		body = accessSection + "\n" + body
	}
	screen := stScreen.Width(innerWidth).Render(body)

	box := stBox.Width(innerWidth + 2).Render(screen)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, box)
}

// renderProgress draws "█████░░░░░░░ 1:23 / 3:45" for the current track, or
// just the elapsed time when the duration is unknown. Empty before the first
// progress event arrives. Paused tracks get an amber ⏸ marker.
func renderProgress(elapsed, duration int, paused bool) string {
	if elapsed == 0 && duration == 0 {
		return ""
	}
	prefix := ""
	if paused {
		prefix = stAmber.Render("⏸ ")
	}
	clock := player.FormatClock(float64(elapsed))
	if duration <= 0 {
		return prefix + stDim.Render(clock)
	}
	const width = 12
	filled := elapsed * width / duration
	if filled > width {
		filled = width
	}
	bar := stAccent.Render(strings.Repeat("█", filled)) + stDim.Render(strings.Repeat("░", width-filled))
	return prefix + bar + " " + stDim.Render(clock+" / "+player.FormatClock(float64(duration)))
}

func renderChart(rows []statView, width int) string {
	if len(rows) == 0 {
		return stDim.Render("no plays yet")
	}
	max := 0
	for _, r := range rows {
		if r.Count > max {
			max = r.Count
		}
	}
	var lines []string
	for i, r := range rows {
		rank := stAmber.Render(fmt.Sprintf("#%d", i+1))
		barLen := 0
		if max > 0 {
			barLen = (r.Count * 10) / max
			if barLen < 1 {
				barLen = 1
			}
		}
		bar := stAccent.Render(strings.Repeat("█", barLen))
		count := stDim.Render(fmt.Sprintf(" %d", r.Count))
		name := r.Name
		// truncate long names
		maxName := width - 18
		if maxName < 6 {
			maxName = 6
		}
		if len([]rune(name)) > maxName {
			name = string([]rune(name)[:maxName-1]) + "…"
		}
		lines = append(lines, rank+" "+stFg.Render(name)+" "+bar+count)
	}
	return strings.Join(lines, "\n")
}

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
