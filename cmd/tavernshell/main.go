package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/angusmclean/tavernshell/core/dice"
	"github.com/angusmclean/tavernshell/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// If arguments provided, run in single-command mode
	if len(os.Args) > 1 {
		runSingleCommand(os.Args[1:])
		return
	}

	// Otherwise, launch interactive TUI
	runInteractive()
}

// runSingleCommand executes a single command and exits
func runSingleCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("No command provided")
		os.Exit(1)
	}

	cmd := strings.ToLower(args[0])

	switch {
	case strings.HasPrefix("roll", cmd) || strings.HasPrefix("r", cmd):
		if len(args) < 2 {
			fmt.Println("Usage: tavernshell roll <dice>")
			fmt.Println("Examples: tavernshell roll 2d6+3, tavernshell roll d20!, tavernshell roll 4d6kh3")
			os.Exit(1)
		}

		notation := args[1]

		// Parse the expression
		expr, err := dice.Parse(notation)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		// Roll the dice
		result, err := dice.RollExpression(expr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("🎲 %s\n", result.String())

	case strings.HasPrefix("help", cmd) || strings.HasPrefix("h", cmd):
		printHelp()

	default:
		// Try to parse the first argument as a dice roll
		notation := args[0]
		expr, err := dice.Parse(notation)
		if err != nil {
			// Not a valid dice roll, show unknown command error
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
			fmt.Fprintln(os.Stderr, "Run 'tavernshell help' for usage information")
			os.Exit(1)
		}

		// It's a valid dice roll! Execute it
		result, err := dice.RollExpression(expr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("🎲 %s\n", result.String())
	}
}

// runInteractive starts the interactive TUI
func runInteractive() {
	p := tea.NewProgram(
		tui.NewModel(),
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// printHelp displays help information
func printHelp() {
	help := `TavernShell - D&D Tools CLI

USAGE:
  tavernshell              Start interactive mode
  tavernshell <command>    Run a single command

COMMANDS:
  roll <dice>   Roll dice with modifiers, advantage, keep/drop
  help          Show this help message

EXAMPLES:
  tavernshell              # Start interactive shell
  tavernshell roll 2d6     # Roll 2 six-sided dice
  tavernshell r d20+5      # Roll d20 and add 5
  tavernshell r d20!       # Roll d20 with advantage
  tavernshell r 4d6kh3     # Roll 4d6, keep highest 3
  tavernshell r 4d6dl1     # Roll 4d6, drop lowest 1

DICE NOTATION:
  XdY       - Roll X dice with Y sides each
  XdY+Z     - Add modifier Z to the total
  XdY!      - Advantage (roll each die twice, keep highest)
  XdYkhN    - Keep highest N dice
  XdYklN    - Keep lowest N dice
  XdYdhN    - Drop highest N dice
  XdYdlN    - Drop lowest N dice

Dropped dice are shown in angle brackets ‹like this›

INTERACTIVE MODE FEATURES:
  - Dice rolling with full notation support
  - Countdown alarms with display
  - Initiative/turn tracking with visual panel
  - Number trackers (HP, AC, etc.) with pinning
  
In interactive mode, type 'h' for help or 'q' to quit.
`
	fmt.Print(help)
}
