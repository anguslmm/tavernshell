# TavernShell ⚔️

A fast, minimal D&D tools CLI for DMs and players. Built with Go for speed and portability.

## Features

- 🎲 **Fast Dice Rolling** - Quick dice rolls with fair RNG (crypto/rand)
- 🖥️ **Interactive TUI** - Beautiful terminal interface powered by Bubble Tea
- ⚡ **Single Command Mode** - Run commands directly without entering the shell
- 🔤 **Smart Shortcuts** - Single-letter commands (e.g., `r` for roll)

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/angusmclean/tavernshell.git
cd tavernshell

# Build
go build -o tavernshell ./cmd/tavernshell

# Run
./tavernshell
```

## Usage

### Interactive Mode

Run without arguments to enter interactive mode:

```bash
./tavernshell
```

Then use commands like:
- `r 2d6` - Roll 2 six-sided dice
- `roll d20` - Roll a d20
- `h` - Show help
- `q` - Quit

### Single Command Mode

Execute commands directly:

```bash
./tavernshell roll 2d6
./tavernshell r d20
```

## Dice Notation

Currently supports simple dice notation:
- `d20` - Roll one 20-sided die
- `2d6` - Roll two 6-sided dice
- `4d8` - Roll four 8-sided dice

More advanced notation (modifiers, advantage/disadvantage) coming soon!

## Architecture

The project is designed with modularity in mind:

- `core/` - Business logic (dice rolling, etc.) - UI-agnostic
- `tui/` - Bubble Tea TUI implementation
- `cmd/` - CLI entry points

This separation makes it easy to build a web API or desktop app in the future.

## Roadmap

- [x] Basic dice rolling (XdY notation)
- [x] Interactive TUI
- [x] Single command mode
- [ ] Advanced dice notation (modifiers, e.g., 2d6+3)
- [ ] Initiative tracker
- [ ] Counters and trackers
- [ ] Advantage/disadvantage for d20
- [ ] State persistence
- [ ] Web API
- [ ] Desktop app

## Development

### Run Tests

```bash
go test ./...
```

### Project Structure

```
tavernshell/
├── cmd/
│   └── tavernshell/     # Main entry point
├── core/
│   └── dice/            # Dice rolling logic
├── tui/                 # Terminal UI
├── go.mod
└── README.md
```

## License

MIT

