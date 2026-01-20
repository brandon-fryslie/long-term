package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// Mode represents the current operating mode
type Mode uint32

const (
	ModeNormal  Mode = 0 // Normal I/O passthrough
	ModeCommand Mode = 1 // Command mode (UI active, intercept input)
)

// KeyCode represents parsed keyboard input
type KeyCode int

const (
	KeyUnknown KeyCode = iota
	KeyChar             // Regular character (a-z, 0-9, space, etc)
	KeyESC
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyBackspace
	KeyEnter
)

// KeyEvent represents a parsed keyboard event
type KeyEvent struct {
	Code      KeyCode
	Char      rune   // Valid when Code == KeyChar
	Shift     bool   // Modifier flags
	Ctrl      bool
	ShiftCtrl bool
}

// NumericMode tracks numeric input state
type NumericMode int

const (
	NumericNone   NumericMode = iota
	NumericHeight             // 'n' pressed - entering absolute height
	NumericDelta              // 'd' pressed - entering delta
)

// NumericBuffer accumulates numeric input
type NumericBuffer struct {
	mode   NumericMode
	digits []rune
}

func (nb *NumericBuffer) reset() {
	nb.mode = NumericNone
	nb.digits = nil
}

func (nb *NumericBuffer) append(r rune) {
	nb.digits = append(nb.digits, r)
}

func (nb *NumericBuffer) backspace() {
	if len(nb.digits) > 0 {
		nb.digits = nb.digits[:len(nb.digits)-1]
	}
}

func (nb *NumericBuffer) value() (int, error) {
	if len(nb.digits) == 0 {
		return 0, fmt.Errorf("empty input")
	}
	s := string(nb.digits)
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	// Validate range
	if nb.mode == NumericHeight {
		if val < 1 || val > 9999 {
			return 0, fmt.Errorf("height must be 1-9999")
		}
	} else if nb.mode == NumericDelta {
		// Delta requires +/- prefix
		if len(s) == 0 || (s[0] != '+' && s[0] != '-') {
			return 0, fmt.Errorf("delta requires +/- prefix")
		}
		if val < -9999 || val > 9999 {
			return 0, fmt.Errorf("delta must be ±1 to ±9999")
		}
	}
	return val, nil
}

// keyboardParser reads stdin and emits KeyEvent structs
type keyboardParser struct {
	eventChan  chan KeyEvent
	buf        []byte
	state      int // 0=idle, 1=saw ESC, 2=saw ESC[
	lastESC    time.Time
	escTimeout time.Duration
}

func newKeyboardParser() *keyboardParser {
	return &keyboardParser{
		eventChan:  make(chan KeyEvent, 10),
		buf:        make([]byte, 0, 16),
		escTimeout: 100 * time.Millisecond,
	}
}

// Write implements io.Writer to observe stdin bytes
func (kp *keyboardParser) Write(p []byte) (n int, err error) {
	for _, b := range p {
		kp.processByte(b)
	}
	return len(p), nil
}

func (kp *keyboardParser) processByte(b byte) {
	now := time.Now()

	// Check for ESC timeout
	if kp.state > 0 && !kp.lastESC.IsZero() && now.Sub(kp.lastESC) > kp.escTimeout {
		// Timeout - emit standalone ESC
		kp.eventChan <- KeyEvent{Code: KeyESC}
		kp.state = 0
		kp.buf = kp.buf[:0]
	}

	switch kp.state {
	case 0: // Idle
		if b == 0x1B { // ESC
			kp.state = 1
			kp.lastESC = now
		} else if b == 0x7F { // Backspace
			kp.eventChan <- KeyEvent{Code: KeyBackspace}
		} else if b == '\r' || b == '\n' {
			kp.eventChan <- KeyEvent{Code: KeyEnter}
		} else if b >= 0x20 && b < 0x7F { // Printable ASCII
			kp.eventChan <- KeyEvent{Code: KeyChar, Char: rune(b)}
		}
		// Ignore other control characters

	case 1: // Saw ESC
		if b == '[' {
			kp.state = 2
			kp.buf = kp.buf[:0]
		} else {
			// Not a sequence, emit ESC and process this byte
			kp.eventChan <- KeyEvent{Code: KeyESC}
			kp.state = 0
			kp.processByte(b) // Reprocess current byte
		}

	case 2: // Saw ESC[
		kp.buf = append(kp.buf, b)
		// Check for complete sequences
		if b >= 0x40 && b <= 0x7E { // Final byte of sequence
			kp.parseSequence()
			kp.state = 0
			kp.buf = kp.buf[:0]
		}
	}
}

func (kp *keyboardParser) parseSequence() {
	seq := string(kp.buf)

	// Parse arrow keys with modifiers
	// Basic arrows: A=up, B=down, C=right, D=left
	// Modified: ESC[1;XY where X=modifier, Y=direction
	// Modifiers: 2=Shift, 5=Ctrl, 6=Shift+Ctrl

	var event KeyEvent

	switch seq {
	case "A":
		event.Code = KeyUp
	case "B":
		event.Code = KeyDown
	case "C":
		event.Code = KeyRight
	case "D":
		event.Code = KeyLeft
	case "1;2A":
		event = KeyEvent{Code: KeyUp, Shift: true}
	case "1;2B":
		event = KeyEvent{Code: KeyDown, Shift: true}
	case "1;5A":
		event = KeyEvent{Code: KeyUp, Ctrl: true}
	case "1;5B":
		event = KeyEvent{Code: KeyDown, Ctrl: true}
	case "1;6A":
		event = KeyEvent{Code: KeyUp, ShiftCtrl: true}
	case "1;6B":
		event = KeyEvent{Code: KeyDown, ShiftCtrl: true}
	default:
		event.Code = KeyUnknown
	}

	if event.Code != KeyUnknown {
		kp.eventChan <- event
	}
}

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

// magicDetector observes stdin for Ctrl+\ pressed 3 times within 500ms
type magicDetector struct {
	toggleChan  chan bool
	lastPress   time.Time
	pressCount  int
	magicByte   byte
	window      time.Duration
	targetCount int
}

func newMagicDetector(toggleChan chan bool) *magicDetector {
	return &magicDetector{
		toggleChan:  toggleChan,
		magicByte:   0x1C, // Ctrl+\ (SIGQUIT character)
		window:      500 * time.Millisecond,
		targetCount: 3,
	}
}

func (m *magicDetector) Write(p []byte) (n int, err error) {
	now := time.Now()

	// Check if we need to reset the counter (window expired)
	if !m.lastPress.IsZero() && now.Sub(m.lastPress) > m.window {
		m.pressCount = 0
	}

	// Count occurrences of magic byte in this chunk
	count := bytes.Count(p, []byte{m.magicByte})
	if count > 0 {
		m.pressCount += count
		m.lastPress = now

		if m.pressCount >= m.targetCount {
			// Trigger toggle
			select {
			case m.toggleChan <- true:
			default:
			}
			m.pressCount = 0 // Reset after triggering
		}
	}

	return len(p), nil
}

func run(args []string, initialHeight, initialDelta int) error {
	// Get the real terminal size
	realWidth, realHeight, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		// Default to 80 if we can't get the size
		realWidth = 80
		realHeight = 24
	}

	// Height and delta: single source of truth (atomic for lock-free access)
	var currentHeight atomic.Int32
	var currentDelta atomic.Int32
	currentHeight.Store(int32(initialHeight))
	currentDelta.Store(int32(initialDelta))

	// Calculate effective height
	effectiveHeight := initialHeight
	if initialDelta != 0 {
		effectiveHeight = realHeight + initialDelta
		if effectiveHeight < 1 {
			effectiveHeight = 1
		}
	}

	// Mode state: single source of truth
	var currentMode atomic.Uint32
	currentMode.Store(uint32(ModeNormal))

	// Toggle mode: false = use fake/delta height, true = use real height
	var useRealSize atomic.Bool
	useRealSize.Store(false)

	// Numeric input state
	var numericBuf NumericBuffer

	// Create keyboard parser
	kbParser := newKeyboardParser()

	// Create SIGWINCH channel first (needed by toggle handler)
	sigwinch := make(chan os.Signal, 1)

	// Channel for entering command mode (Ctrl+\ x3)
	enterCommandChan := make(chan bool, 1)
	go func() {
		for range enterCommandChan {
			currentMode.Store(uint32(ModeCommand))
			// TODO: Show UI in Sprint 2
		}
	}()

	// Command handler goroutine - processes keyboard events in command mode
	go func() {
		for event := range kbParser.eventChan {
			if Mode(currentMode.Load()) != ModeCommand {
				continue
			}

			// Handle numeric input mode
			if numericBuf.mode != NumericNone {
				switch event.Code {
				case KeyESC:
					numericBuf.reset()
					// TODO: Update UI
				case KeyBackspace:
					numericBuf.backspace()
					// TODO: Update UI
				case KeyEnter:
					val, err := numericBuf.value()
					if err == nil {
						// Apply value
						if numericBuf.mode == NumericHeight {
							currentHeight.Store(int32(val))
							currentDelta.Store(0) // Clear delta when setting absolute
						} else if numericBuf.mode == NumericDelta {
							currentDelta.Store(int32(val))
						}
						numericBuf.reset()
						// Trigger resize
						sigwinch <- syscall.SIGWINCH
					}
					// TODO: Show error if invalid
				case KeyChar:
					if event.Char >= '0' && event.Char <= '9' {
						numericBuf.append(event.Char)
					} else if numericBuf.mode == NumericDelta && len(numericBuf.digits) == 0 {
						if event.Char == '+' || event.Char == '-' {
							numericBuf.append(event.Char)
						}
					}
					// TODO: Update UI
				}
				continue
			}

			// Normal command mode handling
			switch event.Code {
			case KeyESC:
				// Exit command mode
				currentMode.Store(uint32(ModeNormal))
				numericBuf.reset()
				// TODO: Clear UI in Sprint 2

			case KeyChar:
				switch event.Char {
				case 'n':
					numericBuf.mode = NumericHeight
					numericBuf.digits = nil
					// TODO: Update UI to show prompt
				case 'd':
					numericBuf.mode = NumericDelta
					numericBuf.digits = nil
					// TODO: Update UI to show prompt
				case ' ':
					// Toggle real/fake size
					current := useRealSize.Load()
					useRealSize.Store(!current)
					sigwinch <- syscall.SIGWINCH
					// TODO: Update UI
				case 'r':
					// Reset to defaults from flags
					currentHeight.Store(int32(initialHeight))
					currentDelta.Store(int32(initialDelta))
					useRealSize.Store(false)
					sigwinch <- syscall.SIGWINCH
					// TODO: Update UI
				}

			case KeyUp, KeyDown:
				// Adjust height based on modifiers
				delta := 1
				if event.Shift {
					delta = 20
				} else if event.Ctrl {
					delta = 200
				} else if event.ShiftCtrl {
					delta = 200
				}
				if event.Code == KeyDown {
					delta = -delta
				}

				// Apply delta
				if currentDelta.Load() != 0 {
					currentDelta.Add(int32(delta))
				} else {
					newHeight := int(currentHeight.Load()) + delta
					if newHeight < 1 {
						newHeight = 1
					} else if newHeight > 9999 {
						newHeight = 9999
					}
					currentHeight.Store(int32(newHeight))
				}
				sigwinch <- syscall.SIGWINCH
				// TODO: Update UI
			}
		}
	}()

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
	signal.Notify(sigwinch, syscall.SIGWINCH)
	go func() {
		for range sigwinch {
			if w, h, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
				// Determine which height to use based on toggle state
				targetHeight := h
				if !useRealSize.Load() {
					// Use fake/delta height
					delta := int(currentDelta.Load())
					if delta != 0 {
						targetHeight = h + delta
						if targetHeight < 1 {
							targetHeight = 1
						}
					} else {
						targetHeight = int(currentHeight.Load())
					}
				}
				// else: use real height (targetHeight already = h)

				pty.Setsize(ptmx, &pty.Winsize{
					Rows: uint16(targetHeight),
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
	// stdin -> pty (with magic key detection and keyboard parsing)
	go func() {
		magicDet := newMagicDetector(enterCommandChan)
		// Chain: stdin -> magicDetector -> keyboardParser -> pty (if normal mode)
		multiWriter := io.MultiWriter(magicDet, kbParser)
		tee := io.TeeReader(os.Stdin, multiWriter)

		// Copy to PTY, but check mode first
		buf := make([]byte, 1024)
		for {
			n, err := tee.Read(buf)
			if err != nil {
				break
			}

			// Only forward to PTY if in normal mode
			if Mode(currentMode.Load()) == ModeNormal {
				ptmx.Write(buf[:n])
			}
			// In command mode, input is intercepted by keyboard parser
		}
	}()
	// pty -> stdout
	go func() {
		io.Copy(os.Stdout, ptmx)
	}()

	// Wait for the command to finish
	return cmd.Wait()
}
