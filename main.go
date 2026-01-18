package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

func main() {
	height := flag.Int("height", 10000, "fake terminal height to report to the wrapped program")
	heightDelta := flag.Int("delta", 2000, "report real_height + delta (positive adds rows, negative subtracts; overrides -height if set)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] -- command [args...]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Wraps a command with a PTY that reports a fake terminal height.\n")
		fmt.Fprintf(os.Stderr, "Width is passed through from the real terminal.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Validate delta usage - must be explicitly positive or negative
	if *heightDelta != 0 {
		// Check if user passed the flag
		deltaSet := false
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "delta" {
				deltaSet = true
			}
		})
		if deltaSet {
			// User set delta, but we can't distinguish between "5" and "+5"
			// So we require explicit sign by checking the original arg
			foundExplicitSign := false
			for i, arg := range os.Args {
				if arg == "-delta" && i+1 < len(os.Args) {
					val := os.Args[i+1]
					if len(val) > 0 && (val[0] == '+' || val[0] == '-') {
						foundExplicitSign = true
						break
					}
				}
			}
			if !foundExplicitSign && *heightDelta > 0 {
				fmt.Fprintf(os.Stderr, "Error: -delta requires explicit sign (use +%d or -%d, not %d)\n", *heightDelta, *heightDelta, *heightDelta)
				os.Exit(1)
			}
		}
	}

	if err := run(args, *height, *heightDelta); err != nil {
		fmt.Fprintf(os.Stderr, "giant-pty: %v\n", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}

func run(args []string, fakeHeight, heightDelta int) error {
	// Get the real terminal size
	realWidth, realHeight, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		// Default to 80 if we can't get the size
		realWidth = 80
		realHeight = 24
	}

	// Calculate effective height
	effectiveHeight := fakeHeight
	if heightDelta != 0 {
		effectiveHeight = realHeight + heightDelta
		if effectiveHeight < 1 {
			effectiveHeight = 1
		}
	}

	// Create the command, using shell if needed for aliases
	var cmd *exec.Cmd
	if _, err := exec.LookPath(args[0]); err != nil {
		// Command not found in PATH, try through shell for aliases
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		cmdStr := strings.Join(args, " ")
		cmd = exec.Command(shell, "-ic", cmdStr)
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}

	// Start with PTY using our effective size
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: uint16(effectiveHeight),
		Cols: uint16(realWidth),
	})
	if err != nil {
		return fmt.Errorf("failed to start pty: %w", err)
	}
	defer ptmx.Close()

	// Handle SIGWINCH (window resize)
	sigwinch := make(chan os.Signal, 1)
	signal.Notify(sigwinch, syscall.SIGWINCH)
	go func() {
		for range sigwinch {
			if w, h, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
				// Recalculate effective height on resize
				h := h
				if heightDelta != 0 {
					h = h + heightDelta
					if h < 1 {
						h = 1
					}
				}
				pty.Setsize(ptmx, &pty.Winsize{
					Rows: uint16(h),
					Cols: uint16(w),
				})
			}
		}
	}()
	// Trigger initial resize
	sigwinch <- syscall.SIGWINCH

	// Put terminal into raw mode (only if stdin is a terminal)
	var oldState *term.State
	if term.IsTerminal(int(os.Stdin.Fd())) {
		oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to set raw mode: %w", err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	}

	// Proxy I/O
	// stdin -> pty
	go func() {
		io.Copy(ptmx, os.Stdin)
	}()
	// pty -> stdout
	go func() {
		io.Copy(os.Stdout, ptmx)
	}()

	// Wait for the command to finish
	return cmd.Wait()
}
