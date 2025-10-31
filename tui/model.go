package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/angusmclean/tavernshell/core/dice"
	"github.com/angusmclean/tavernshell/core/tracker/number"
	"github.com/angusmclean/tavernshell/core/tracker/rotation"
	"github.com/angusmclean/tavernshell/core/tracker/timer"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxHistory = 100

// tickMsg is sent every second to update timers
type tickMsg time.Time

// Model represents the TUI application state
type Model struct {
	textInput            textinput.Model   // text input component
	history              []string          // command history/results (displayed output)
	commandHistory       []string          // command history (for up/down arrow navigation)
	historyIndex         int               // current position in command history (-1 = not navigating)
	timerManager         *timer.Manager    // manages active timers
	initiativeManager    *rotation.Manager // manages initiative/rotation tracker
	numberTrackerManager *number.Manager   // manages number trackers
	width                int               // terminal width
	height               int               // terminal height
	initiativeEntryMode  bool              // true when entering initiative participants
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
		textInput:            ti,
		history:              []string{"Welcome to TavernShell! Type 'h' for help."},
		commandHistory:       []string{},
		historyIndex:         -1,
		timerManager:         timer.NewManager(),
		initiativeManager:    rotation.NewManager(),
		numberTrackerManager: number.NewManager(),
		initiativeEntryMode:  false,
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
				m.addHistory(fmt.Sprintf("⏰ Alarm '%s' finished (%s)", t.Label, timer.FormatDuration(t.Duration)))
			} else {
				m.addHistory(fmt.Sprintf("⏰ Alarm finished (%s)", timer.FormatDuration(t.Duration)))
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
	// Special handling for initiative entry mode
	if m.initiativeEntryMode {
		// Check for exit commands
		if input == "" || strings.ToLower(input) == "done" || strings.ToLower(input) == "end" {
			m.initiativeEntryMode = false
			m.addHistory("Initiative setup complete. Use 'i n' to advance turns.")
			return nil
		}
		// Parse "name initiative" format
		parts := strings.Fields(input)
		if len(parts) < 2 {
			m.addHistory("Format: <name> <initiative> (or 'done' to finish)")
			return nil
		}
		initiative, err := strconv.Atoi(parts[len(parts)-1])
		if err != nil {
			m.addHistory("Invalid initiative value. Format: <name> <initiative>")
			return nil
		}
		name := strings.Join(parts[:len(parts)-1], " ")
		m.initiativeManager.Add(name, initiative)
		m.addHistory(fmt.Sprintf("Added %s (initiative %d)", name, initiative))
		return nil
	}

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
	case strings.HasPrefix("alarm", cmd) || cmd == "a":
		m.handleTimer(parts[1:])
		return nil
	case strings.HasPrefix("initiative", cmd) || strings.HasPrefix("init", cmd) || cmd == "i":
		m.handleInitiative(parts[1:])
		return nil
	case strings.HasPrefix("tracker", cmd) || strings.HasPrefix("track", cmd) || cmd == "t":
		m.handleTrack(parts[1:])
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
		// Try to parse the entire input as a dice roll
		expr, err := dice.Parse(input)
		if err != nil {
			// Not a valid dice roll, show unknown command error
			m.addHistory(fmt.Sprintf("Unknown command: %s (type 'h' for help)", cmd))
			return nil
		}

		// It's a valid dice roll! Execute it
		result, err := dice.RollExpression(expr)
		if err != nil {
			m.addHistory(fmt.Sprintf("Error: %s", err))
			return nil
		}

		// Format and display the result
		m.addHistory(fmt.Sprintf("🎲 %s", formatDiceResult(result)))
		return nil
	}
}

// handleRoll processes a roll command
func (m *Model) handleRoll(args []string) {
	if len(args) == 0 {
		m.addHistory("Usage: r/roll <dice> (e.g., 'r 2d6+3', 'r d20!', 'r 4d6kh3')")
		return
	}

	notation := args[0]

	// Parse the expression
	expr, err := dice.Parse(notation)
	if err != nil {
		m.addHistory(fmt.Sprintf("Error: %s", err))
		return
	}

	// Roll the dice
	result, err := dice.RollExpression(expr)
	if err != nil {
		m.addHistory(fmt.Sprintf("Error: %s", err))
		return
	}

	// Format and display the result with styling
	m.addHistory(fmt.Sprintf("🎲 %s", formatDiceResult(result)))
}

// handleTimer processes an alarm command
func (m *Model) handleTimer(args []string) {
	if len(args) == 0 {
		m.addHistory("Usage: a/alarm <duration> [name] (e.g., 'a 5m', 'a 1h concentration')")
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
		m.addHistory(fmt.Sprintf("⏰ Started alarm '%s' for %s", label, timer.FormatDuration(duration)))
	} else {
		m.addHistory(fmt.Sprintf("⏰ Started alarm for %s", timer.FormatDuration(duration)))
	}
}

// handleInitiative processes initiative commands
func (m *Model) handleInitiative(args []string) {
	if len(args) == 0 {
		m.addHistory("Usage: i/init <command> - Commands: start/s, next/n, add/a, kill/k, end/e")
		return
	}

	subCmd := strings.ToLower(args[0])

	switch {
	case strings.HasPrefix("start", subCmd) || subCmd == "s":
		m.initiativeManager.Start()
		m.initiativeEntryMode = true
		m.addHistory("Starting initiative. Enter '<name> <initiative>' for each participant.")
		m.addHistory("Type 'done' when finished.")

	case strings.HasPrefix("next", subCmd) || subCmd == "n":
		if !m.initiativeManager.IsActive() {
			m.addHistory("No active initiative. Use 'i start' to begin.")
			return
		}
		m.initiativeManager.Next()
		tracker := m.initiativeManager.GetTracker()
		if tracker != nil {
			current := tracker.GetCurrent()
			if current != nil {
				m.addHistory(fmt.Sprintf("Turn: %s (Initiative %d) - Round %d", current.Name, current.Initiative, tracker.Round))
			}
		}

	case strings.HasPrefix("add", subCmd) || subCmd == "a":
		if !m.initiativeManager.IsActive() {
			m.addHistory("No active initiative. Use 'i start' to begin.")
			return
		}
		m.initiativeEntryMode = true
		m.addHistory("Enter '<name> <initiative>' or 'done' to finish.")

	case strings.HasPrefix("kill", subCmd) || subCmd == "k":
		if !m.initiativeManager.IsActive() {
			m.addHistory("No active initiative.")
			return
		}
		if len(args) < 2 {
			m.addHistory("Usage: i kill <name>")
			return
		}
		name := strings.Join(args[1:], " ")
		err := m.initiativeManager.MarkOut(name)
		if err != nil {
			m.addHistory(fmt.Sprintf("Error: %s", err))
		} else {
			m.addHistory(fmt.Sprintf("%s is out of combat", name))
		}

	case strings.HasPrefix("end", subCmd) || subCmd == "e":
		m.initiativeManager.End()
		m.addHistory("Initiative ended.")

	default:
		m.addHistory(fmt.Sprintf("Unknown initiative command: %s", subCmd))
	}
}

// handleTrack processes tracker commands
func (m *Model) handleTrack(args []string) {
	if len(args) == 0 {
		m.addHistory("Usage: t/tracker <command> - Commands: add/a, set/s, adjust/adj, list/l, pin/p, unpin/u, pinall/pa, delete/d, deleteall/da, search/f")
		return
	}

	subCmd := strings.ToLower(args[0])

	switch {
	case strings.HasPrefix("add", subCmd) || subCmd == "a":
		if len(args) < 4 {
			m.addHistory("Usage: track add <name> <current> <max>")
			return
		}
		name := args[1]
		current, err1 := strconv.Atoi(args[2])
		max, err2 := strconv.Atoi(args[3])
		if err1 != nil || err2 != nil {
			m.addHistory("Error: current and max must be numbers")
			return
		}
		m.numberTrackerManager.Add(name, current, max)
		m.addHistory(fmt.Sprintf("Added tracker: [%s] %d/%d", name, current, max))

	case strings.HasPrefix("set", subCmd) || subCmd == "s":
		if len(args) < 3 {
			m.addHistory("Usage: track set <name> <value>")
			return
		}
		name := args[1]
		value, err := strconv.Atoi(args[2])
		if err != nil {
			m.addHistory("Error: value must be a number")
			return
		}
		tracker := m.numberTrackerManager.Get(name)
		if tracker == nil {
			m.addHistory(fmt.Sprintf("Tracker '%s' not found", name))
			return
		}
		tracker.Set(value)
		m.addHistory(fmt.Sprintf("[%s] %d/%d", tracker.Name, tracker.Current, tracker.Max))

	case strings.HasPrefix("adjust", subCmd) || subCmd == "adj":
		if len(args) < 3 {
			m.addHistory("Usage: track adjust <name> <delta> (e.g., '+5' or '-10')")
			return
		}
		name := args[1]
		delta, err := strconv.Atoi(args[2])
		if err != nil {
			m.addHistory("Error: delta must be a number (e.g., +5 or -10)")
			return
		}
		tracker := m.numberTrackerManager.Get(name)
		if tracker == nil {
			m.addHistory(fmt.Sprintf("Tracker '%s' not found", name))
			return
		}
		tracker.Adjust(delta)
		m.addHistory(fmt.Sprintf("[%s] %d/%d", tracker.Name, tracker.Current, tracker.Max))

	case strings.HasPrefix("list", subCmd) || subCmd == "l":
		trackers := m.numberTrackerManager.List()
		if len(trackers) == 0 {
			m.addHistory("No trackers")
			return
		}
		m.addHistory("Trackers:")
		for _, t := range trackers {
			pinned := ""
			if t.Pinned {
				pinned = " (pinned)"
			}
			m.addHistory(fmt.Sprintf("  [%s] %d/%d%s", t.Name, t.Current, t.Max, pinned))
		}

	case strings.HasPrefix("pin", subCmd) || subCmd == "p":
		if len(args) < 2 {
			m.addHistory("Usage: track pin <name>")
			return
		}
		name := args[1]
		tracker := m.numberTrackerManager.Get(name)
		if tracker == nil {
			m.addHistory(fmt.Sprintf("Tracker '%s' not found", name))
			return
		}
		tracker.Pin()
		m.addHistory(fmt.Sprintf("Pinned [%s]", tracker.Name))

	case strings.HasPrefix("unpin", subCmd) || subCmd == "u":
		if len(args) < 2 {
			m.addHistory("Usage: track unpin <name>")
			return
		}
		name := args[1]
		tracker := m.numberTrackerManager.Get(name)
		if tracker == nil {
			m.addHistory(fmt.Sprintf("Tracker '%s' not found", name))
			return
		}
		tracker.Unpin()
		m.addHistory(fmt.Sprintf("Unpinned [%s]", tracker.Name))

	case strings.HasPrefix("pinall", subCmd) || subCmd == "pa":
		count := m.numberTrackerManager.PinAll()
		m.addHistory(fmt.Sprintf("Pinned %d tracker(s)", count))

	case strings.HasPrefix("delete", subCmd) || subCmd == "d":
		if len(args) < 2 {
			m.addHistory("Usage: track delete <name>")
			return
		}
		name := args[1]
		err := m.numberTrackerManager.Delete(name)
		if err != nil {
			m.addHistory(fmt.Sprintf("Error: %s", err))
		} else {
			m.addHistory(fmt.Sprintf("Deleted tracker '%s'", name))
		}

	case strings.HasPrefix("deleteall", subCmd) || subCmd == "da":
		m.numberTrackerManager.DeleteAll()
		m.addHistory("Deleted all trackers")

	case strings.HasPrefix("search", subCmd) || subCmd == "f":
		if len(args) < 2 {
			m.addHistory("Usage: track search <pattern>")
			return
		}
		pattern := args[1]
		results := m.numberTrackerManager.Search(pattern)
		if len(results) == 0 {
			m.addHistory(fmt.Sprintf("No trackers matching '%s'", pattern))
			return
		}
		m.addHistory(fmt.Sprintf("Trackers matching '%s':", pattern))
		for _, t := range results {
			pinned := ""
			if t.Pinned {
				pinned = " (pinned)"
			}
			m.addHistory(fmt.Sprintf("  [%s] %d/%d%s", t.Name, t.Current, t.Max, pinned))
		}

	default:
		m.addHistory(fmt.Sprintf("Unknown track command: %s", subCmd))
	}
}

// handleHelp shows available commands
func (m *Model) handleHelp() {
	help := []string{
		"Available Commands:",
		"  r/roll <dice>           - Roll dice with modifiers, advantage, keep/drop",
		"  a/alarm <time> [name]   - Start a countdown alarm (e.g., 'a 5m', 'a 1h concentration')",
		"  i/init <cmd>            - Initiative tracking (start/s, next/n, add/a, kill/k, end/e)",
		"  t/tracker <cmd>         - Number trackers (add/a, set/s, adjust/adj, list/l, pin/p, etc.)",
		"  h/help                  - Show this help message",
		"  c/clear                 - Clear history",
		"  q/quit                  - Exit (or press Ctrl+C/Esc)",
		"",
		"Dice Examples:",
		"  r 2d6                   - Roll 2 six-sided dice",
		"  r d20+5                 - Roll d20 and add 5",
		"  r d20!                  - Roll d20 with advantage (roll twice, keep highest)",
		"  r 4d6kh3                - Roll 4d6, keep highest 3",
		"",
		"Alarm Examples:",
		"  a 5m                    - Start a 5-minute alarm",
		"  a 1h concentration      - Start a 1-hour alarm named 'concentration'",
		"",
		"Initiative Examples:",
		"  i start                 - Start initiative entry (or 'i s')",
		"  i add                   - Add more participants (or 'i a')",
		"  i next                  - Advance to next turn (or 'i n')",
		"  i kill Goblin           - Mark Goblin as out of combat (or 'i k')",
		"  i end                   - End initiative (or 'i e')",
		"",
		"Tracker Examples:",
		"  t add HP 35 45          - Create HP tracker at 35/45 (or 't a HP 35 45')",
		"  t set HP 40             - Set HP to 40 (or 't s HP 40')",
		"  t adjust HP -10         - Subtract 10 from HP (or 't adj HP -10')",
		"  t pin HP                - Pin HP to top display (or 't p HP')",
		"  t unpin HP              - Unpin HP from display (or 't u HP')",
		"  t list                  - List all trackers (or 't l')",
		"  t pinall                - Pin all trackers (or 't pa')",
		"  t delete HP             - Delete HP tracker (or 't d HP')",
		"  t deleteall             - Delete all trackers (or 't da')",
		"  t search HP             - Search for trackers (or 't f HP')",
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

// buildTrackerBar builds a horizontal display of pinned trackers with progress bars
func (m Model) buildTrackerBar() string {
	pinnedTrackers := m.numberTrackerManager.GetPinned()
	if len(pinnedTrackers) == 0 {
		return ""
	}

	trackerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("cyan")).
		Width(0)

	numTrackers := len(pinnedTrackers)
	if numTrackers == 0 {
		return ""
	}

	// Calculate width per tracker slot
	separatorWidth := 3 // " | "
	availableWidth := m.width - ((numTrackers - 1) * separatorWidth)
	if availableWidth < 30 {
		availableWidth = 30 // minimum width
	}
	slotWidth := availableWidth / numTrackers

	var parts []string

	for _, tracker := range pinnedTrackers {
		// Calculate percentage filled
		percentFilled := 0.0
		if tracker.Max > 0 {
			percentFilled = (float64(tracker.Current) / float64(tracker.Max)) * 100.0
			if percentFilled > 100.0 {
				percentFilled = 100.0
			}
			if percentFilled < 0.0 {
				percentFilled = 0.0
			}
		}

		// Build the tracker display
		// Format: [name] current/max [bar]
		icon := fmt.Sprintf("[%s]", tracker.Name)
		valueStr := fmt.Sprintf("%d/%d", tracker.Current, tracker.Max)

		// Calculate available space for bar
		// icon + space + valueStr + space + [bar]
		baseWidth := len(icon) + 1 + len(valueStr) + 1 + 2 // +2 for []

		barWidth := slotWidth - baseWidth
		if barWidth < 10 {
			barWidth = 10
			// If we need more space, truncate name
			maxNameLen := slotWidth - 2 - 1 - len(valueStr) - 1 - barWidth - 2 // -2 for [], -1 for spaces
			if maxNameLen < 0 {
				maxNameLen = 0
			}
			if len(tracker.Name) > maxNameLen {
				var truncName string
				if maxNameLen > 3 {
					truncName = tracker.Name[:maxNameLen-3] + "..."
				} else if maxNameLen > 0 {
					truncName = tracker.Name[:maxNameLen]
				} else {
					truncName = ""
				}
				icon = fmt.Sprintf("[%s]", truncName)
			}
		}

		// Build progress bar (empty at 0, full at max)
		filled := int(percentFilled / 100.0 * float64(barWidth))
		if filled < 0 {
			filled = 0
		}
		if filled > barWidth {
			filled = barWidth
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

		// Build the text
		trackerText := fmt.Sprintf("%s %s [%s]", icon, valueStr, bar)

		// Pad to slot width
		textLen := len(icon) + 1 + len(valueStr) + 1 + 1 + len(bar) + 1 // spaces + [ + bar + ]
		if textLen < slotWidth {
			trackerText += strings.Repeat(" ", slotWidth-textLen)
		}

		parts = append(parts, trackerStyle.Render(trackerText))
	}

	return strings.Join(parts, " | ")
}

// buildInitiativePanel builds the right-side initiative panel
func (m Model) buildInitiativePanel() []string {
	if !m.initiativeManager.IsActive() {
		return nil
	}

	tracker := m.initiativeManager.GetTracker()
	if tracker == nil || !tracker.HasParticipants() {
		return nil
	}

	var lines []string

	// Round header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("green"))
	lines = append(lines, headerStyle.Render(fmt.Sprintf("Round %d", tracker.Round)))
	lines = append(lines, strings.Repeat("─", 25))

	// Participants
	currentStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("yellow")).
		Background(lipgloss.Color("236"))

	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))

	inactiveStyle := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("241"))

	for i, p := range tracker.Participants {
		var line string
		isCurrent := (i == tracker.CurrentTurn)

		// Format: "  Name (init)"
		text := fmt.Sprintf("  %s (%d)", p.Name, p.Initiative)

		// Truncate if too long
		if len(text) > 25 {
			text = text[:22] + "..."
		}

		if !p.IsActive {
			// Inactive/dead
			line = inactiveStyle.Render(text + " ✗")
		} else if isCurrent {
			// Current turn
			line = currentStyle.Render("▶ " + text[2:])
		} else {
			// Active but not current
			line = activeStyle.Render(text)
		}

		lines = append(lines, line)
	}

	return lines
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

	// Build tracker bar
	trackerBar := m.buildTrackerBar()

	// Build initiative panel
	initiativePanel := m.buildInitiativePanel()

	// Calculate widths for initiative panel
	initiativePanelWidth := 28
	hasInitiative := len(initiativePanel) > 0
	mainWidth := m.width
	if hasInitiative {
		mainWidth = m.width - initiativePanelWidth - 1 // -1 for separator
	}

	// Build the input line with help text
	inputLine := promptStyle.Render("➤ ") + m.textInput.View()
	helpText := helpStyle.Render("  Ctrl+C or 'q' to quit")

	// Calculate available height for history
	headerLines := 2       // title + separator
	timerTrackerLines := 2 // timer bar + blank line
	if trackerBar != "" {
		timerTrackerLines += 4 // blank line + separator + tracker bar + separator
	}
	footerLines := 2 // input + help
	availableHeight := m.height - headerLines - timerTrackerLines - footerLines

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
	b.WriteString("\n")

	// Tracker bar (if any pinned trackers)
	if trackerBar != "" {
		b.WriteString("\n")
		b.WriteString(strings.Repeat("─", m.width))
		b.WriteString("\n")
		b.WriteString(trackerBar)
		b.WriteString("\n")
		b.WriteString(strings.Repeat("─", m.width))
		b.WriteString("\n")
	} else {
		b.WriteString("\n")
	}

	// Main content area - split if initiative is active
	if hasInitiative {
		// Calculate padding to push history to bottom
		currentLines := headerLines + timerTrackerLines + len(historyLines)
		paddingLines := m.height - currentLines - footerLines

		// Content lines to display
		contentLines := make([]string, availableHeight)

		// Add padding
		for i := 0; i < paddingLines && i < availableHeight; i++ {
			contentLines[i] = ""
		}

		// Add history
		historyStart := paddingLines
		if historyStart < 0 {
			historyStart = 0
		}
		for i, line := range historyLines {
			idx := historyStart + i
			if idx < availableHeight {
				contentLines[idx] = line
			}
		}

		// Combine main content with initiative panel
		for i := 0; i < availableHeight; i++ {
			mainLine := contentLines[i]
			// Pad or truncate main line to mainWidth
			if len(mainLine) > mainWidth {
				mainLine = mainLine[:mainWidth]
			} else if len(mainLine) < mainWidth {
				mainLine += strings.Repeat(" ", mainWidth-len(mainLine))
			}

			// Get initiative panel line if available
			initLine := ""
			if i < len(initiativePanel) {
				initLine = initiativePanel[i]
			}

			b.WriteString(mainLine)
			b.WriteString(" ")
			b.WriteString(initLine)
			b.WriteString("\n")
		}
	} else {
		// No initiative - simple layout
		currentLines := headerLines + timerTrackerLines + len(historyLines)
		paddingLines := m.height - currentLines - footerLines

		// Add padding at top to push history down
		if paddingLines > 0 {
			b.WriteString(strings.Repeat("\n", paddingLines))
		}

		// History area (appears right above input, growing upward)
		if len(historyLines) > 0 {
			b.WriteString(strings.Join(historyLines, "\n"))
			b.WriteString("\n")
		}
	}

	// Input line at bottom
	b.WriteString(inputLine)
	b.WriteString("\n")
	b.WriteString(helpText)

	return b.String()
}

// formatDiceResult formats a dice result with styled output for dropped dice
func formatDiceResult(r *dice.Result) string {
	if r == nil {
		return "<nil result>"
	}

	var b strings.Builder

	// Faint style for dropped dice
	faintStyle := lipgloss.NewStyle().Faint(true)

	// Build the notation
	notation := fmt.Sprintf("%dd%d", r.Expression.Count, r.Expression.Sides)
	if r.Expression.Advantage {
		notation += "!"
	}
	if r.Expression.Operation != nil {
		notation += formatOperation(r.Expression.Operation)
	}
	if r.Expression.Modifier != 0 {
		if r.Expression.Modifier > 0 {
			notation += fmt.Sprintf("+%d", r.Expression.Modifier)
		} else {
			notation += fmt.Sprintf("%d", r.Expression.Modifier)
		}
	}

	b.WriteString(notation)
	b.WriteString(": [")

	// Show all dice with styling for dropped ones
	diceStrs := make([]string, len(r.Rolls))
	for i, die := range r.Rolls {
		if die.Kept {
			diceStrs[i] = fmt.Sprintf("%d", die.Value)
		} else {
			// Use faint styling for dropped dice
			diceStrs[i] = faintStyle.Render(fmt.Sprintf("‹%d›", die.Value))
		}
	}
	b.WriteString(strings.Join(diceStrs, ", "))
	b.WriteString("]")

	// Show modifier if present
	if r.Expression.Modifier != 0 {
		if r.Expression.Modifier > 0 {
			b.WriteString(fmt.Sprintf(" +%d", r.Expression.Modifier))
		} else {
			b.WriteString(fmt.Sprintf(" %d", r.Expression.Modifier))
		}
	}

	// Show total
	b.WriteString(fmt.Sprintf(" = %d", r.Total))

	// Add description if there are dropped dice
	hasDropped := false
	for _, die := range r.Rolls {
		if !die.Kept {
			hasDropped = true
			break
		}
	}

	if hasDropped {
		b.WriteString(" ")
		b.WriteString(faintStyle.Render("("))
		if r.Expression.Advantage {
			b.WriteString(faintStyle.Render("advantage"))
		} else if r.Expression.Operation != nil {
			b.WriteString(faintStyle.Render(formatOperationDescription(r.Expression.Operation)))
		}
		b.WriteString(faintStyle.Render(")"))
	}

	return b.String()
}

// formatOperation formats an operation for display in notation
func formatOperation(op *dice.Operation) string {
	switch op.Type {
	case dice.OpKeepHighest:
		return fmt.Sprintf("kh%d", op.Count)
	case dice.OpKeepLowest:
		return fmt.Sprintf("kl%d", op.Count)
	case dice.OpDropHighest:
		return fmt.Sprintf("dh%d", op.Count)
	case dice.OpDropLowest:
		return fmt.Sprintf("dl%d", op.Count)
	default:
		return ""
	}
}

// formatOperationDescription returns a human-readable description of an operation
func formatOperationDescription(op *dice.Operation) string {
	switch op.Type {
	case dice.OpKeepHighest:
		return fmt.Sprintf("kept highest %d", op.Count)
	case dice.OpKeepLowest:
		return fmt.Sprintf("kept lowest %d", op.Count)
	case dice.OpDropHighest:
		return fmt.Sprintf("dropped highest %d", op.Count)
	case dice.OpDropLowest:
		return fmt.Sprintf("dropped lowest %d", op.Count)
	default:
		return ""
	}
}
