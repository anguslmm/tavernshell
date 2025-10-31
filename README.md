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

**Download a pre-built binary** from the [releases page](https://github.com/yourusername/tavernshell/releases), or build from source:

```bash
git clone https://github.com/yourusername/tavernshell.git
cd tavernshell
go build -o tavernshell ./cmd/tavernshell
./tavernshell
```

Or if you have Go installed:

```bash
go install github.com/yourusername/tavernshell/cmd/tavernshell@latest
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
- `roll 2d6` or `r 2d6` - Roll dice
- `d20` - Roll a single d20 (shorthand)

**Number Trackers:**
- `track hp 45` - Create a tracker named "hp" starting at 45
- `hp -10` - Adjust by -10 (take damage)
- `hp +5` - Adjust by +5 (heal)
- `hp 45` - Set to specific value
- `list` or `ls` - Show all trackers

**Timers:**
- `timer bless 10` - Start 10-round countdown for "bless"
- `timer haste 10r` - Same thing (r for rounds)
- `list` - See active timers (updates automatically)

**Turn Rotation:**
- `rotation goblin wizard fighter` - Create turn order
- `next` or `n` - Advance to next turn
- `list` - See current turn and upcoming

**General:**
- `help` or `h` - Show help
- `quit` or `q` - Exit

## Dice Notation

Supports standard D&D dice notation:
- `d20` - Single die
- `2d6` - Multiple dice
- `d20+5` - Roll with modifiers
- `3d8-2` - Negative modifiers too

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

