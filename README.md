# giant-pty

A long-term PTY wrapper that allows you to run commands with a fake terminal height, useful for testing and scripts that need specific terminal dimensions.

## Features

- Wraps a command with a PTY that reports a fake terminal height
- Supports fixed height or delta-based height adjustments
- **Interactive command mode** for runtime height control
- Terminal width is passed through from the real terminal
- Responsive to terminal resize events (SIGWINCH)
- Unix/macOS only

## Installation

### Quick Install

```bash
curl -fsSL https://github.com/bmf/giant-pty/releases/latest/download/$(uname -s | tr A-Z a-z)-$(uname -m) -o ~/.local/bin/long-term && chmod +x ~/.local/bin/long-term
```

### From GitHub Releases

Download the latest binary for your platform from the [releases page](https://github.com/bmf/giant-pty/releases):

#### macOS

```bash
# For Apple Silicon (M1/M2/etc)
curl -L https://github.com/bmf/giant-pty/releases/download/v0.1.0/long-term-darwin-arm64 -o long-term
chmod +x long-term
sudo mv long-term /usr/local/bin/

# For Intel Macs
curl -L https://github.com/bmf/giant-pty/releases/download/v0.1.0/long-term-darwin-amd64 -o long-term
chmod +x long-term
sudo mv long-term /usr/local/bin/
```

#### Linux

```bash
# For x86_64
curl -L https://github.com/bmf/giant-pty/releases/download/v0.1.0/long-term-linux-amd64 -o long-term
chmod +x long-term
sudo mv long-term /usr/local/bin/

# For ARM64
curl -L https://github.com/bmf/giant-pty/releases/download/v0.1.0/long-term-linux-arm64 -o long-term
chmod +x long-term
sudo mv long-term /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/bmf/giant-pty.git
cd giant-pty
go build -o long-term .
sudo mv long-term /usr/local/bin/
```

Or use the justfile:

```bash
just install  # Builds and installs to ~/.local/bin/
```

## Usage

```
giant-pty [flags] -- command [args...]
```

### Flags

- `-height` (default: 10000): Report this fake terminal height to the wrapped program
- `-delta` (default: 0): Report real_height + delta (use explicit sign, e.g., +2000 or -500; overrides -height if set)

### Examples

```bash
# Report a fixed height of 50 rows
long-term -height 50 -- bash

# Report 20 rows more than the real terminal height
long-term -delta +20 -- vim

# Report 10 rows less than the real terminal height
long-term -delta -10 -- tmux
```

## Interactive Command Mode

Press **Ctrl+\\** three times (within 500ms) to enter interactive command mode. A UI overlay will appear showing:

```
┌──────────────────────────────────────┐
│   LONG-TERM ENABLED                  │
├──────────────────────────────────────┤
│ Term size: 80x100 (Δ+20)             │
│                                      │
│ UP/DOWN: ±1  Shift: ±20  Ctrl: ±200  │
│ n: set height  d: set delta          │
│ space: toggle  r: reset  ESC: exit   │
└──────────────────────────────────────┘
```

### Command Mode Controls

**Arrow Keys:**
- **UP/DOWN**: Adjust height by ±1
- **Shift+UP/DOWN**: Adjust by ±20
- **Ctrl+UP/DOWN** or **Shift+Ctrl+UP/DOWN**: Adjust by ±200

**Numeric Entry:**
- **n**: Enter absolute height (1-9999)
- **d**: Enter delta offset (±1 to ±9999, requires +/- prefix)

**Other Commands:**
- **Space**: Toggle between fake and real terminal height
- **r**: Reset to original command-line flags
- **ESC**: Exit command mode

### Command Mode Notes

- The UI overlay refreshes every 100ms while active
- Input to the wrapped process is paused during command mode
- Wrapped process output continues to scroll (UI stays overlaid)
- Terminal resize events update the UI position
- Command mode requires `/dev/tty` access (unavailable in piped contexts)

## How It Works

`giant-pty` creates a pseudo-terminal (PTY) wrapper around your command with a configurable reported height. This is useful for:

- Testing terminal applications with specific height constraints
- Running interactive programs that check terminal dimensions
- Working with tools that have height-based behavior differences
- Dynamically adjusting terminal height without restarting applications

The actual terminal width is always passed through from your real terminal, and the wrapper responds to terminal resize events.

## License

MIT
