package diffview

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type LineKind int

const (
	LineUnchanged LineKind = iota
	LineAdded
	LineRemoved
)

type DiffLine struct {
	LeftNo    int
	RightNo   int
	LeftText  string
	RightText string
	Kind      LineKind
}

type Options struct {
	Width           int
	Height          int
	Title           string
	LeftTitle       string
	RightTitle      string
	ShowLineNumbers bool
}

type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Help     key.Binding
	Quit     key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("up/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("down/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("pgup/b", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "f"),
			key.WithHelp("pgdn/f", "page down"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "bottom"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.PageUp, k.PageDown, k.Help, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.PageUp, k.PageDown}, {k.Top, k.Bottom, k.Help, k.Quit}}
}

type Model struct {
	vp         viewport.Model
	help       help.Model
	keys       KeyMap
	showHelp   bool
	options    Options
	diffLines  []DiffLine
	leftLines  int
	rightLines int
}

func New(opts Options) Model {
	if opts.Width <= 0 {
		opts.Width = 100
	}
	if opts.Height <= 0 {
		opts.Height = 28
	}
	if opts.Title == "" {
		opts.Title = "Diff View"
	}
	if opts.LeftTitle == "" {
		opts.LeftTitle = "left"
	}
	if opts.RightTitle == "" {
		opts.RightTitle = "right"
	}

	vp := viewport.New(opts.Width, opts.Height-2)
	vp.SetContent("No diff data yet.")

	h := help.New()
	h.ShowAll = false

	return Model{
		vp:       vp,
		help:     h,
		keys:     DefaultKeyMap(),
		options:  opts,
		showHelp: true,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetDiffStrings(left, right string) {
	m.diffLines, m.leftLines, m.rightLines = buildDiff(left, right)
	m.vp.SetContent(m.renderContent())
	m.vp.GotoTop()
}

func (m *Model) SetDiffLines(lines []DiffLine) {
	m.diffLines = lines
	m.leftLines = 0
	m.rightLines = 0
	for _, line := range lines {
		if line.LeftNo > m.leftLines {
			m.leftLines = line.LeftNo
		}
		if line.RightNo > m.rightLines {
			m.rightLines = line.RightNo
		}
	}
	m.vp.SetContent(m.renderContent())
	m.vp.GotoTop()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.options.Width = msg.Width
		m.options.Height = msg.Height
		m.vp.Width = msg.Width
		bodyHeight := msg.Height - 2
		if !m.showHelp {
			bodyHeight++
		}
		if bodyHeight < 1 {
			bodyHeight = 1
		}
		m.vp.Height = bodyHeight
		m.vp.SetContent(m.renderContent())
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up):
			m.vp.LineUp(1)
		case key.Matches(msg, m.keys.Down):
			m.vp.LineDown(1)
		case key.Matches(msg, m.keys.PageUp):
			m.vp.HalfViewUp()
		case key.Matches(msg, m.keys.PageDown):
			m.vp.HalfViewDown()
		case key.Matches(msg, m.keys.Top):
			m.vp.GotoTop()
		case key.Matches(msg, m.keys.Bottom):
			m.vp.GotoBottom()
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			m.help.ShowAll = m.showHelp
			bodyHeight := m.options.Height - 2
			if !m.showHelp {
				bodyHeight++
			}
			if bodyHeight < 1 {
				bodyHeight = 1
			}
			m.vp.Height = bodyHeight
		}
	}

	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	subtle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	header := fmt.Sprintf("%s  %s -> %s", m.options.Title, m.options.LeftTitle, m.options.RightTitle)
	stats := fmt.Sprintf("left:%d right:%d hunks:%d", m.leftLines, m.rightLines, len(m.diffLines))

	parts := []string{headerStyle.Render(header), subtle.Render(stats), m.vp.View()}
	if m.showHelp {
		parts = append(parts, subtle.Render(m.help.View(m.keys)))
	}
	return strings.Join(parts, "\n")
}

func buildDiff(left, right string) ([]DiffLine, int, int) {
	dmp := diffmatchpatch.New()
	leftChars, rightChars, lineArray := dmp.DiffLinesToChars(left, right)
	diffs := dmp.DiffMain(leftChars, rightChars, false)
	diffs = dmp.DiffCharsToLines(diffs, lineArray)

	leftNo := 1
	rightNo := 1
	lines := make([]DiffLine, 0, len(diffs)*2)

	for _, d := range diffs {
		chunkLines := splitLinesKeepOrder(d.Text)
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			for _, line := range chunkLines {
				if line == "" {
					continue
				}
				lines = append(lines, DiffLine{
					LeftNo:    leftNo,
					RightNo:   rightNo,
					LeftText:  line,
					RightText: line,
					Kind:      LineUnchanged,
				})
				leftNo++
				rightNo++
			}
		case diffmatchpatch.DiffDelete:
			for _, line := range chunkLines {
				if line == "" {
					continue
				}
				lines = append(lines, DiffLine{
					LeftNo:   leftNo,
					RightNo:  0,
					LeftText: line,
					Kind:     LineRemoved,
				})
				leftNo++
			}
		case diffmatchpatch.DiffInsert:
			for _, line := range chunkLines {
				if line == "" {
					continue
				}
				lines = append(lines, DiffLine{
					LeftNo:    0,
					RightNo:   rightNo,
					RightText: line,
					Kind:      LineAdded,
				})
				rightNo++
			}
		}
	}

	return lines, leftNo - 1, rightNo - 1
}

func splitLinesKeepOrder(s string) []string {
	if s == "" {
		return nil
	}
	raw := strings.Split(s, "\n")
	if len(raw) > 0 && raw[len(raw)-1] == "" {
		raw = raw[:len(raw)-1]
	}
	return raw
}

func (m Model) renderContent() string {
	if len(m.diffLines) == 0 {
		return "No changes"
	}

	removedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	addedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	unchangedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	leftWidth := digits(m.leftLines)
	rightWidth := digits(m.rightLines)

	var b strings.Builder
	for i, line := range m.diffLines {
		if i > 0 {
			b.WriteByte('\n')
		}

		sign := " "
		text := line.LeftText
		if line.Kind == LineAdded {
			sign = "+"
			text = line.RightText
		}
		if line.Kind == LineRemoved {
			sign = "-"
		}

		leftStr := padLineNo(line.LeftNo, leftWidth)
		rightStr := padLineNo(line.RightNo, rightWidth)
		if m.options.ShowLineNumbers {
			b.WriteString(numberStyle.Render(leftStr))
			b.WriteString(" | ")
			b.WriteString(numberStyle.Render(rightStr))
			b.WriteString(" ")
		}

		lineOut := sign + " " + text
		switch line.Kind {
		case LineAdded:
			b.WriteString(addedStyle.Render(lineOut))
		case LineRemoved:
			b.WriteString(removedStyle.Render(lineOut))
		default:
			b.WriteString(unchangedStyle.Render(lineOut))
		}
	}
	return b.String()
}

func digits(n int) int {
	if n <= 0 {
		return 1
	}
	return len(strconv.Itoa(n))
}

func padLineNo(n, width int) string {
	if n <= 0 {
		return strings.Repeat(" ", width)
	}
	s := strconv.Itoa(n)
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}
