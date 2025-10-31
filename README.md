# TavernShell

A lightweight terminal tool for D&D sessions. Roll dice, track numbers, manage timers—all without leaving your terminal.

## Features

- Dice rolling with standard notation (2d6, d20, etc.)
- Number trackers for HP, spell slots, or anything else
- Countdown timers for spell durations and effects
- Initiative/turn rotation tracking
- Interactive TUI or single-command mode
- Uses crypto/rand for fair dice rolls

## Installation

**Download a pre-built binary** from the [releases page](https://github.com/anguslmm/tavernshell/releases).

**macOS users:** After downloading, you'll need to make it executable and remove the quarantine flag:
```bash
chmod +x tavernshell-darwin-arm64  # or tavernshell-darwin-amd64 for Intel Macs
xattr -d com.apple.quarantine tavernshell-darwin-arm64
./tavernshell-darwin-arm64
```

**Or if you have Go installed:**

```bash
go install github.com/anguslmm/tavernshell/cmd/tavernshell@latest
```

**Or build from source:**

```bash
git clone https://github.com/anguslmm/tavernshell.git
cd tavernshell
go build -o tavernshell ./cmd/tavernshell
./tavernshell
```

## Quick Start

Run `tavernshell` to enter interactive mode, or pass commands directly:

```bash
# Interactive mode
./tavernshell

# Single commands
./tavernshell roll 2d6
./tavernshell r d20+5
```

### Commands

**Dice Rolling:**
- `r 2d6` or `roll 2d6` - Roll dice
- `2d6` - Just type the notation directly
- `r d20+5` - Roll with modifiers
- `r d20!` - Roll with advantage (keep highest)
- `r 4d6kh3` - Roll 4d6, keep highest 3

**Alarms/Timers:**
- `a 5m` or `alarm 5m` - Start a 5-minute countdown
- `a 30s boulder_hits` - Sometimes players need pressure

**Initiative Tracking:**
- `i start` or `i s` - Begin initiative (then enter `name initiative` for each)
- `i next` or `i n` - Advance to next turn
- `i add` or `i a` - Add more participants
- `i kill Goblin` or `i k Goblin` - Mark as out of combat
- `i end` or `i e` - End initiative

**Number Trackers:**
- `t add HP 35 45` or `t a HP 35 45` - Create tracker at 35/45
- `t set HP 40` or `t s HP 40` - Set to 40
- `t adjust HP -10` or `t adj HP -10` - Adjust by -10
- `t unpin HP` or `t u HP` - Pin to display
- `t pin HP` or `t p HP` - Pin to display
- `t list` or `t l` - Show all trackers
- `t delete HP` or `t d HP` - Delete tracker

**General:**
- `h` or `help` - Show help
- `c` or `clear` - Clear history
- `q` or `quit` - Exit

## Dice Notation

Supports D&D dice notation with modifiers and advanced rolling:
- `d20`, `2d6` - Basic rolls
- `d20+5`, `3d8-2` - Modifiers
- `d20!` - Advantage (roll twice, keep highest)
- `4d6kh3` - Roll 4d6, keep highest 3 (for ability scores)
- `3d6kl2` - Keep lowest 2

## Why?

I wanted a fast way to roll dice and track things during D&D sessions without alt-tabbing to a browser or phone. Plus Go compiles to a single binary, so it's easy to share.

## Development

Run tests:
```bash
go test ./...
```

The codebase is split into `core/` (business logic) and `tui/` (terminal interface), so you could build a web version or GUI on top of the same core if you wanted to.

## License

MIT - see LICENSE file

