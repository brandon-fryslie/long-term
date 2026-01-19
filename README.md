# giant-pty

A long-term PTY wrapper that allows you to run commands with a fake terminal height, useful for testing and scripts that need specific terminal dimensions.

## Features

- Wraps a command with a PTY that reports a fake terminal height
- Supports fixed height or delta-based height adjustments
- Terminal width is passed through from the real terminal
- Responsive to terminal resize events (SIGWINCH)
- Unix/macOS only

## Installation

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

## How It Works

`giant-pty` creates a pseudo-terminal (PTY) wrapper around your command with a configurable reported height. This is useful for:

- Testing terminal applications with specific height constraints
- Running interactive programs that check terminal dimensions
- Working with tools that have height-based behavior differences

The actual terminal width is always passed through from your real terminal, and the wrapper responds to terminal resize events.

## License

MIT
