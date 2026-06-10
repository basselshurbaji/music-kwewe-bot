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
	stBold   = lipgloss.NewStyle().Foreground(colAccent).Bold(true)

	stBox = lipgloss.NewStyle().
		Background(colBg).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colBorder)

	stTitleBar = lipgloss.NewStyle().
			Background(lipgloss.Color("#0e141d")).
			Foreground(colDim).
			PaddingLeft(1).PaddingRight(1)

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
	Title   string `json:"title"`
	AddedBy string `json:"added_by"`
	URL     string `json:"url"`
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
type errMsg struct{ err error }
type tickMsg time.Time

// ── model ─────────────────────────────────────────────────────────────────────

type model struct {
	apiURL  string
	state   stateView
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
	return model{apiURL: "http://localhost" + port + "/api/state"}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchCmd(m.apiURL), tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		return m, tea.Batch(fetchCmd(m.apiURL), tickCmd())
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

// ── view ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	innerWidth := m.width - 4 // 2 box border + 2 padding each side
	if innerWidth < 40 {
		innerWidth = 40
	}

	// Title bar
	dots := "● ● ●"
	title := stDim.Render("  music-kwewe-bot — dashboard (read-only)")
	titleBar := stTitleBar.Width(innerWidth).Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f56")).Render("●"),
			" ",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#ffbd2e")).Render("●"),
			" ",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#27c93f")).Render("●"),
			stDim.Render("  music-kwewe-bot — dashboard (read-only)"),
		),
	)
	_ = dots
	_ = title

	// Prompt line
	prompt := stDim.Render("music-kwewe-bot") + stDim.Render(" ~ $ ") + stBold.Render("status") + stAccent.Render("█")

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
	screen := stScreen.Width(innerWidth).Render(
		prompt + "\n" +
			nowSection + "\n" +
			queueSection + "\n" +
			statsSection + "\n" +
			footer,
	)

	// Outer box
	content := lipgloss.JoinVertical(lipgloss.Left, titleBar, screen)
	box := stBox.Width(innerWidth + 2).Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, box)
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
