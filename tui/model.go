package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/angusmclean/tavernshell/core/dice"
	"github.com/angusmclean/tavernshell/core/timer"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxHistory = 20

// tickMsg is sent every second to update timers
type tickMsg time.Time

// Model represents the TUI application state
type Model struct {
	textInput      textinput.Model // text input component
	history        []string        // command history/results (displayed output)
	commandHistory []string        // command history (for up/down arrow navigation)
	historyIndex   int             // current position in command history (-1 = not navigating)
	timerManager   *timer.Manager  // manages active timers
	width          int             // terminal width
	height         int             // terminal height
}

// NewModel creates a new TUI model
func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50 // Will be updated on first window size message
	ti.Prompt = ""
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

	return Model{
		textInput:      ti,
		history:        []string{"Welcome to TavernShell! Type 'h' for help."},
		commandHistory: []string{},
		historyIndex:   -1,
		timerManager:   timer.NewManager(),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// tickCmd returns a command that sends a tick message every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		// Check for expired timers
		expired := m.timerManager.GetExpired()
		for _, t := range expired {
			if t.Label != "" {
				m.addHistory(fmt.Sprintf("⏰ Timer '%s' expired (%s)", t.Label, timer.FormatDuration(t.Duration)))
			} else {
				m.addHistory(fmt.Sprintf("⏰ Timer expired (%s)", timer.FormatDuration(t.Duration)))
			}
		}
		// Return another tick command to keep updating
		return m, tickCmd()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			input := m.textInput.Value()
			if input != "" {
				// Add to command history
				m.commandHistory = append(m.commandHistory, input)
				if len(m.commandHistory) > maxHistory {
					m.commandHistory = m.commandHistory[1:]
				}

				cmd := m.handleCommand(input)
				m.textInput.Reset()
				m.historyIndex = -1 // Reset history navigation
				return m, cmd
			}

		case tea.KeyUp:
			// Navigate backward in command history
			if len(m.commandHistory) > 0 {
				if m.historyIndex == -1 {
					// Start from the most recent command
					m.historyIndex = len(m.commandHistory) - 1
					m.textInput.SetValue(m.commandHistory[m.historyIndex])
				} else if m.historyIndex > 0 {
					// Go to older command
					m.historyIndex--
					m.textInput.SetValue(m.commandHistory[m.historyIndex])
				}
			}
			return m, nil

		case tea.KeyDown:
			// Navigate forward in command history
			if m.historyIndex != -1 {
				if m.historyIndex < len(m.commandHistory)-1 {
					// Go to newer command
					m.historyIndex++
					m.textInput.SetValue(m.commandHistory[m.historyIndex])
				} else {
					// At the newest command, clear input
					m.historyIndex = -1
					m.textInput.Reset()
				}
			}
			return m, nil

		default:
			// Let textinput handle all other keys (left, right, backspace, characters, etc.)
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// handleCommand processes a command and updates history
func (m *Model) handleCommand(input string) tea.Cmd {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	cmd := strings.ToLower(parts[0])

	// Support single-letter shortcuts
	switch {
	case strings.HasPrefix("roll", cmd):
		m.handleRoll(parts[1:])
		return nil
	case strings.HasPrefix("timer", cmd):
		m.handleTimer(parts[1:])
		return nil
	case strings.HasPrefix("help", cmd):
		m.handleHelp()
		return nil
	case strings.HasPrefix("quit", cmd):
		return tea.Quit
	case strings.HasPrefix("clear", cmd):
		m.history = []string{"Welcome to TavernShell! Type 'h' for help."}
		return nil
	default:
		m.addHistory(fmt.Sprintf("Unknown command: %s (type 'h' for help)", cmd))
		return nil
	}
}

// handleRoll processes a roll command
func (m *Model) handleRoll(args []string) {
	if len(args) == 0 {
		m.addHistory("Usage: r/roll <dice> (e.g., 'r 2d6' or 'roll d20')")
		return
	}

	notation := args[0]
	result, err := dice.RollDice(notation)
	if err != nil {
		m.addHistory(fmt.Sprintf("Error: %s", err))
		return
	}

	m.addHistory(fmt.Sprintf("🎲 %s", result.String()))
}

// handleTimer processes a timer command
func (m *Model) handleTimer(args []string) {
	if len(args) == 0 {
		m.addHistory("Usage: t/timer <duration> [name] (e.g., 't 5m', 't 1h concentration')")
		return
	}

	durationStr := args[0]
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		m.addHistory(fmt.Sprintf("Error: invalid duration '%s' (use format like: 1h, 5m, 30s, 1h30m)", durationStr))
		return
	}

	if duration <= 0 {
		m.addHistory("Error: duration must be positive")
		return
	}

	// Optional label is everything after the duration
	label := ""
	if len(args) > 1 {
		label = strings.Join(args[1:], " ")
	}

	newTimer := timer.NewTimer(duration, label)
	m.timerManager.Add(newTimer)

	if label != "" {
		m.addHistory(fmt.Sprintf("⏱  Started timer '%s' for %s", label, timer.FormatDuration(duration)))
	} else {
		m.addHistory(fmt.Sprintf("⏱  Started timer for %s", timer.FormatDuration(duration)))
	}
}

// handleHelp shows available commands
func (m *Model) handleHelp() {
	help := []string{
		"Available Commands:",
		"  r/roll <dice>         - Roll dice (e.g., 'r 2d6', 'roll d20')",
		"  t/timer <time> [name] - Start a timer (e.g., 't 5m', 't 1h concentration')",
		"  h/help                - Show this help message",
		"  c/clear               - Clear history",
		"  q/quit                - Exit (or press Ctrl+C/Esc)",
		"",
		"Examples:",
		"  r 2d6                 - Roll 2 six-sided dice",
		"  roll d20              - Roll a twenty-sided die",
		"  t 5m                  - Start a 5-minute timer",
		"  t 1h concentration    - Start a 1-hour timer named 'concentration'",
	}
	for _, line := range help {
		m.addHistory(line)
	}
}

// addHistory adds a line to the history
func (m *Model) addHistory(line string) {
	m.history = append(m.history, line)
	if len(m.history) > maxHistory {
		m.history = m.history[len(m.history)-maxHistory:]
	}
}

// buildTimerBar builds a horizontal display of 3 timer slots spanning the window width
func (m Model) buildTimerBar() string {
	activeTimers := m.timerManager.GetActive()

	timerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Width(0) // Don't let lipgloss add extra width

	emptyStyle := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("241")).
		Width(0)

	// Calculate width per timer slot
	separatorWidth := 3                              // " | "
	availableWidth := m.width - (2 * separatorWidth) // space for 2 separators
	if availableWidth < 30 {
		availableWidth = 30 // minimum width
	}
	slotWidth := availableWidth / 3

	var parts []string

	// Always show 3 timer slots
	for i := 0; i < 3; i++ {
		if i < len(activeTimers) {
			t := activeTimers[i]
			remaining := t.Remaining()
			percentRemaining := 100.0 - t.PercentComplete() // Countdown: 100% at start, 0% at end

			// Build the timer display
			// Format: [T] label time [bar]
			// Use [T] instead of emoji to avoid width issues
			icon := "[T]"
			timeStr := fmt.Sprintf("%s/%s",
				timer.FormatDurationShort(remaining),
				timer.FormatDurationShort(t.Duration))

			// Calculate available space
			// icon (3) + space (1) + timeStr + space (1) + [bar]
			baseWidth := len(icon) + 1 + len(timeStr) + 1 + 2 // +2 for []

			// Determine bar width - aim for at least 10 chars
			labelLen := len(t.Label)
			if labelLen > 0 {
				baseWidth += labelLen + 1 // +1 for space
			}

			barWidth := slotWidth - baseWidth
			if barWidth < 10 {
				barWidth = 10
				// If we need more space, truncate label
				if labelLen > 0 {
					maxLabelLen := slotWidth - len(icon) - 1 - len(timeStr) - 1 - barWidth - 2 - 1
					if maxLabelLen < 0 {
						maxLabelLen = 0
					}
					if labelLen > maxLabelLen {
						labelLen = maxLabelLen
					}
				}
			}

			// Truncate label if needed
			label := t.Label
			if len(label) > labelLen {
				if labelLen > 3 {
					label = label[:labelLen-3] + "..."
				} else if labelLen > 0 {
					label = label[:labelLen]
				} else {
					label = ""
				}
			}

			// Build countdown progress bar (full at start, empty at end)
			filled := int(percentRemaining / 100.0 * float64(barWidth))
			if filled < 0 {
				filled = 0
			}
			if filled > barWidth {
				filled = barWidth
			}
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

			// Build the text
			var timerText string
			if label != "" {
				timerText = fmt.Sprintf("%s %s %s [%s]", icon, label, timeStr, bar)
			} else {
				timerText = fmt.Sprintf("%s %s [%s]", icon, timeStr, bar)
			}

			// Pad to slot width
			textLen := len(icon) + 1 + len(label)
			if label != "" {
				textLen += 1 // space after label
			}
			textLen += len(timeStr) + 1 + 1 + len(bar) + 1 // space + [ + bar + ]

			if textLen < slotWidth {
				timerText += strings.Repeat(" ", slotWidth-textLen)
			}

			parts = append(parts, timerStyle.Render(timerText))
		} else {
			// Empty slot - pad to slot width
			emptyText := "[T] (empty)"
			textLen := len(emptyText)
			if textLen < slotWidth {
				emptyText += strings.Repeat(" ", slotWidth-textLen)
			}
			parts = append(parts, emptyStyle.Render(emptyText))
		}
	}

	return strings.Join(parts, " | ")
}

// View renders the TUI
func (m Model) View() string {
	if m.height == 0 {
		return "Loading..."
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	promptStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))

	helpStyle := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("241"))

	// Build the title bar
	titleBar := titleStyle.Render("⚔️  TavernShell")

	// Build timer display (horizontal)
	timerBar := m.buildTimerBar()

	// Build the input line with help text
	inputLine := promptStyle.Render("➤ ") + m.textInput.View()
	helpText := helpStyle.Render("  Ctrl+C or 'q' to quit")

	// Calculate available height for history
	// -1 for title, -1 for separator, -1 for timer bar (always shown), -1 for blank line, -1 for input, -1 for help
	timerLines := 2 // timer bar + blank line (always shown)
	availableHeight := m.height - 5 - timerLines

	// Get the history lines to display (most recent at bottom)
	var historyLines []string
	if len(m.history) > 0 {
		start := 0
		if len(m.history) > availableHeight {
			start = len(m.history) - availableHeight
		}
		historyLines = m.history[start:]
	}

	// Build the full view
	var b strings.Builder

	// Title
	b.WriteString(titleBar)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")

	// Timer bar (always shown)
	b.WriteString(timerBar)
	b.WriteString("\n\n")

	// Calculate padding to push history to bottom (right above input)
	currentLines := 2 + timerLines + len(historyLines) // title + separator + timers + history
	paddingLines := m.height - currentLines - 2        // -2 for input and help

	// Add padding at top to push history down
	if paddingLines > 0 {
		b.WriteString(strings.Repeat("\n", paddingLines))
	}

	// History area (appears right above input, growing upward)
	if len(historyLines) > 0 {
		b.WriteString(strings.Join(historyLines, "\n"))
	}

	// Input line at bottom
	b.WriteString("\n")
	b.WriteString(inputLine)
	b.WriteString("\n")
	b.WriteString(helpText)

	return b.String()
}
