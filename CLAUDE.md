# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

long-term (binary name: `long-term`) is a PTY wrapper that reports a fake terminal height to wrapped programs. Unix/macOS only.

## Build Commands

```bash
just build          # Build binary (outputs ./long-term)
just install        # Build and install to ~/.local/bin/
go build -o long-term .  # Direct Go build
```

## Testing

No test suite exists. Manual testing:
```bash
./long-term -height 50 -- bash -c 'tput lines'  # Should print 50
./long-term -delta +20 -- bash -c 'tput lines'  # Should print real height + 20
./test-toggle.sh  # Interactive test for runtime toggling (see below)
```

## Release Process

```bash
just release patch  # Bump patch version, tag, and push (triggers GHA release)
just release minor  # Bump minor version
just release major  # Bump major version
```

GitHub Actions builds cross-platform binaries (linux/darwin × amd64/arm64) on tag push.

## Architecture

Single-file Go program (`main.go`, ~220 lines):

1. **Argument parsing**: `-height` for fixed height, `-delta` for relative adjustment (requires explicit +/- sign)
2. **PTY creation**: Uses `creack/pty` to spawn command with fake dimensions
3. **Signal handling**: SIGWINCH handler updates PTY size on terminal resize (recalculates delta mode)
4. **I/O proxying**: Bidirectional copy between stdin/stdout and PTY in raw mode
5. **Runtime toggle**: `io.TeeReader` observes stdin for magic key sequence (Ctrl+\ x3 within 500ms) to toggle between fake and real terminal height

Key behavior: Width always passes through from real terminal; only height is faked.

### Runtime Size Toggle

The wrapper can toggle between fake and real terminal height at runtime:

- **Magic sequence**: Press Ctrl+\ three times within 500ms
- **Implementation**: `io.TeeReader` observes stdin stream without consuming bytes; passes all input through unchanged to wrapped process
- **Use case**: Applications that need large terminal most of the time but real size for specific screens

The toggle mechanism uses an atomic bool flag that the SIGWINCH handler checks when recalculating PTY dimensions. The magic key detector counts occurrences of byte 0x1C (Ctrl+\) and triggers a toggle when threshold is met within time window.
