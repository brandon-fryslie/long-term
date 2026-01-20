# Test Case Templates & Implementation Guide

Quick reference for implementing the 74 unit tests and 12 integration tests identified in the audit.

---

## Unit Test Template Structure

### File: `keyboard_parser_test.go`

```go
package main

import (
	"testing"
	"time"
)

// TestKeyboardParser wraps newKeyboardParser with test-friendly channel reading
type TestKeyboardParser struct {
	*keyboardParser
	events []KeyEvent
}

func newTestKeyboardParser() *TestKeyboardParser {
	kp := newKeyboardParser()
	return &TestKeyboardParser{keyboardParser: kp, events: []KeyEvent{}}
}

// readAllEvents drains eventChan with timeout
func (tkp *TestKeyboardParser) readAllEvents(t *testing.T, timeoutMs int) []KeyEvent {
	ticker := time.NewTicker(time.Duration(timeoutMs) * time.Millisecond)
	defer ticker.Stop()

	var result []KeyEvent
	for {
		select {
		case evt := <-tkp.eventChan:
			result = append(result, evt)
		case <-ticker.C:
			return result
		}
	}
}

// ============================================================================
// Basic ASCII Tests
// ============================================================================

func TestKeyChar_Regular(t *testing.T) {
	kp := newTestKeyboardParser()
	kp.processByte('a')

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Code != KeyChar || events[0].Char != 'a' {
		t.Errorf("Expected KeyChar('a'), got %v", events[0])
	}
}

func TestKeyChar_Digit(t *testing.T) {
	kp := newTestKeyboardParser()
	kp.processByte('5')

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyChar || events[0].Char != '5' {
		t.Errorf("Expected KeyChar('5'), got %v", events)
	}
}

func TestKeyChar_Space(t *testing.T) {
	kp := newTestKeyboardParser()
	kp.processByte(' ')

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyChar || events[0].Char != ' ' {
		t.Errorf("Expected KeyChar(' '), got %v", events)
	}
}

// ============================================================================
// Special Keys
// ============================================================================

func TestKeyBackspace(t *testing.T) {
	kp := newTestKeyboardParser()
	kp.processByte(0x7F) // Backspace

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyBackspace {
		t.Errorf("Expected KeyBackspace, got %v", events)
	}
}

func TestKeyEnter_CR(t *testing.T) {
	kp := newTestKeyboardParser()
	kp.processByte('\r')

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyEnter {
		t.Errorf("Expected KeyEnter (CR), got %v", events)
	}
}

func TestKeyEnter_LF(t *testing.T) {
	kp := newTestKeyboardParser()
	kp.processByte('\n')

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyEnter {
		t.Errorf("Expected KeyEnter (LF), got %v", events)
	}
}

// ============================================================================
// Arrow Keys (Basic)
// ============================================================================

func TestKeyUp(t *testing.T) {
	kp := newTestKeyboardParser()
	// ESC[A = arrow up
	kp.processByte(0x1B) // ESC
	kp.processByte('[')
	kp.processByte('A')

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyUp {
		t.Errorf("Expected KeyUp, got %v", events)
	}
}

func TestKeyDown(t *testing.T) {
	kp := newTestKeyboardParser()
	// ESC[B = arrow down
	kp.processByte(0x1B)
	kp.processByte('[')
	kp.processByte('B')

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyDown {
		t.Errorf("Expected KeyDown, got %v", events)
	}
}

func TestKeyLeft(t *testing.T) {
	kp := newTestKeyboardParser()
	// ESC[D = arrow left
	kp.processByte(0x1B)
	kp.processByte('[')
	kp.processByte('D')

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyLeft {
		t.Errorf("Expected KeyLeft, got %v", events)
	}
}

func TestKeyRight(t *testing.T) {
	kp := newTestKeyboardParser()
	// ESC[C = arrow right
	kp.processByte(0x1B)
	kp.processByte('[')
	kp.processByte('C')

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyRight {
		t.Errorf("Expected KeyRight, got %v", events)
	}
}

// ============================================================================
// Arrow Keys with Modifiers
// ============================================================================

func TestKeyUp_Shift(t *testing.T) {
	kp := newTestKeyboardParser()
	// ESC[1;2A = Shift+Up
	kp.processByte(0x1B)
	kp.processByte('[')
	for _, b := range []byte("1;2A") {
		kp.processByte(b)
	}

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyUp || !events[0].Shift {
		t.Errorf("Expected KeyUp+Shift, got %v", events)
	}
}

func TestKeyDown_Shift(t *testing.T) {
	kp := newTestKeyboardParser()
	// ESC[1;2B = Shift+Down
	kp.processByte(0x1B)
	kp.processByte('[')
	for _, b := range []byte("1;2B") {
		kp.processByte(b)
	}

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyDown || !events[0].Shift {
		t.Errorf("Expected KeyDown+Shift, got %v", events)
	}
}

func TestKeyUp_Ctrl(t *testing.T) {
	kp := newTestKeyboardParser()
	// ESC[1;5A = Ctrl+Up
	kp.processByte(0x1B)
	kp.processByte('[')
	for _, b := range []byte("1;5A") {
		kp.processByte(b)
	}

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyUp || !events[0].Ctrl {
		t.Errorf("Expected KeyUp+Ctrl, got %v", events)
	}
}

func TestKeyDown_Ctrl(t *testing.T) {
	kp := newTestKeyboardParser()
	// ESC[1;5B = Ctrl+Down
	kp.processByte(0x1B)
	kp.processByte('[')
	for _, b := range []byte("1;5B") {
		kp.processByte(b)
	}

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyDown || !events[0].Ctrl {
		t.Errorf("Expected KeyDown+Ctrl, got %v", events)
	}
}

func TestKeyUp_ShiftCtrl(t *testing.T) {
	kp := newTestKeyboardParser()
	// ESC[1;6A = Shift+Ctrl+Up
	kp.processByte(0x1B)
	kp.processByte('[')
	for _, b := range []byte("1;6A") {
		kp.processByte(b)
	}

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyUp || !events[0].ShiftCtrl {
		t.Errorf("Expected KeyUp+ShiftCtrl, got %v", events)
	}
}

func TestKeyDown_ShiftCtrl(t *testing.T) {
	kp := newTestKeyboardParser()
	// ESC[1;6B = Shift+Ctrl+Down
	kp.processByte(0x1B)
	kp.processByte('[')
	for _, b := range []byte("1;6B") {
		kp.processByte(b)
	}

	events := kp.readAllEvents(t, 10)
	if len(events) != 1 || events[0].Code != KeyDown || !events[0].ShiftCtrl {
		t.Errorf("Expected KeyDown+ShiftCtrl, got %v", events)
	}
}

// ============================================================================
// Edge Cases & Timing
// ============================================================================

func TestKeyESC_Standalone(t *testing.T) {
	kp := newTestKeyboardParser()
	// ESC without following [, after timeout
	kp.processByte(0x1B)
	time.Sleep(110 * time.Millisecond) // Wait for timeout (100ms in code)
	kp.processByte('a') // Next key to trigger timeout check

	events := kp.readAllEvents(t, 50)
	// Should emit standalone ESC, then KeyChar('a')
	if len(events) < 1 || events[0].Code != KeyESC {
		t.Errorf("Expected standalone ESC event, got %v", events)
	}
}

func TestKeyMultiple_Sequential(t *testing.T) {
	kp := newTestKeyboardParser()
	kp.processByte('a')
	kp.processByte('b')
	kp.processByte('c')

	events := kp.readAllEvents(t, 10)
	if len(events) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(events))
	}
	expected := []rune{'a', 'b', 'c'}
	for i, exp := range expected {
		if events[i].Code != KeyChar || events[i].Char != exp {
			t.Errorf("Event %d: expected KeyChar('%c'), got %v", i, exp, events[i])
		}
	}
}

func TestKeyUnknown_Ignored(t *testing.T) {
	kp := newTestKeyboardParser()
	kp.processByte(0x00) // NUL
	kp.processByte(0x01) // SOH
	kp.processByte(0x02) // STX

	events := kp.readAllEvents(t, 10)
	if len(events) != 0 {
		t.Errorf("Expected no events for control chars, got %v", events)
	}
}

func TestKeyStateRecovery(t *testing.T) {
	kp := newTestKeyboardParser()
	// Bad escape sequence, then good one
	kp.processByte(0x1B)
	kp.processByte('x') // Invalid, not '['
	time.Sleep(20 * time.Millisecond)
	// Now send valid arrow
	kp.processByte(0x1B)
	kp.processByte('[')
	kp.processByte('A')

	events := kp.readAllEvents(t, 50)
	// Should have standalone ESC, then KeyUp
	if len(events) < 2 {
		t.Fatalf("Expected 2+ events, got %d: %v", len(events), events)
	}
	// Last event should be KeyUp
	if events[len(events)-1].Code != KeyUp {
		t.Errorf("Expected final KeyUp, got %v", events)
	}
}

// Add 9 more edge case tests as needed for full 26-test suite
// (abbreviated for brevity in this template)
```

---

## Numeric Buffer Test Template

### File: `numeric_buffer_test.go`

```go
package main

import (
	"testing"
)

func TestHeightValue_Valid_Min(t *testing.T) {
	nb := &NumericBuffer{mode: NumericHeight}
	nb.append('1')
	val, err := nb.value()
	if err != nil || val != 1 {
		t.Errorf("Expected 1, got val=%d err=%v", val, err)
	}
}

func TestHeightValue_Valid_Max(t *testing.T) {
	nb := &NumericBuffer{mode: NumericHeight}
	for _, ch := range "9999" {
		nb.append(ch)
	}
	val, err := nb.value()
	if err != nil || val != 9999 {
		t.Errorf("Expected 9999, got val=%d err=%v", val, err)
	}
}

func TestHeightValue_Invalid_Zero(t *testing.T) {
	nb := &NumericBuffer{mode: NumericHeight}
	nb.append('0')
	_, err := nb.value()
	if err == nil {
		t.Error("Expected error for height=0")
	}
}

func TestHeightValue_Invalid_Over(t *testing.T) {
	nb := &NumericBuffer{mode: NumericHeight}
	for _, ch := range "10000" {
		nb.append(ch)
	}
	_, err := nb.value()
	if err == nil {
		t.Error("Expected error for height=10000")
	}
}

func TestHeightValue_Empty(t *testing.T) {
	nb := &NumericBuffer{mode: NumericHeight}
	_, err := nb.value()
	if err == nil {
		t.Error("Expected error for empty buffer")
	}
}

func TestDeltaValue_Valid_Plus(t *testing.T) {
	nb := &NumericBuffer{mode: NumericDelta}
	for _, ch := range "+50" {
		nb.append(ch)
	}
	val, err := nb.value()
	if err != nil || val != 50 {
		t.Errorf("Expected +50, got val=%d err=%v", val, err)
	}
}

func TestDeltaValue_Valid_Minus(t *testing.T) {
	nb := &NumericBuffer{mode: NumericDelta}
	for _, ch := range "-50" {
		nb.append(ch)
	}
	val, err := nb.value()
	if err != nil || val != -50 {
		t.Errorf("Expected -50, got val=%d err=%v", val, err)
	}
}

func TestDeltaValue_Invalid_NoSign(t *testing.T) {
	nb := &NumericBuffer{mode: NumericDelta}
	for _, ch := range "50" {
		nb.append(ch)
	}
	_, err := nb.value()
	if err == nil {
		t.Error("Expected error for delta without sign")
	}
}

func TestDeltaValue_Invalid_Over(t *testing.T) {
	nb := &NumericBuffer{mode: NumericDelta}
	for _, ch := range "+10000" {
		nb.append(ch)
	}
	_, err := nb.value()
	if err == nil {
		t.Error("Expected error for delta=+10000")
	}
}

func TestDeltaValue_Valid_Max(t *testing.T) {
	nb := &NumericBuffer{mode: NumericDelta}
	for _, ch := range "+9999" {
		nb.append(ch)
	}
	val, err := nb.value()
	if err != nil || val != 9999 {
		t.Errorf("Expected +9999, got val=%d err=%v", val, err)
	}
}

func TestDeltaValue_Valid_Min(t *testing.T) {
	nb := &NumericBuffer{mode: NumericDelta}
	for _, ch := range "-9999" {
		nb.append(ch)
	}
	val, err := nb.value()
	if err != nil || val != -9999 {
		t.Errorf("Expected -9999, got val=%d err=%v", val, err)
	}
}

func TestAppend_Single(t *testing.T) {
	nb := &NumericBuffer{}
	nb.append('5')
	if len(nb.digits) != 1 || nb.digits[0] != '5' {
		t.Errorf("Expected ['5'], got %v", nb.digits)
	}
}

func TestAppend_Multiple(t *testing.T) {
	nb := &NumericBuffer{}
	for _, ch := range "12345" {
		nb.append(ch)
	}
	if len(nb.digits) != 5 {
		t.Fatalf("Expected 5 digits, got %d", len(nb.digits))
	}
	expected := "12345"
	actual := string(nb.digits)
	if actual != expected {
		t.Errorf("Expected %q, got %q", expected, actual)
	}
}

func TestBackspace_Single(t *testing.T) {
	nb := &NumericBuffer{}
	nb.append('5')
	nb.backspace()
	if len(nb.digits) != 0 {
		t.Errorf("Expected empty, got %v", nb.digits)
	}
}

func TestBackspace_Multiple(t *testing.T) {
	nb := &NumericBuffer{}
	for _, ch := range "123" {
		nb.append(ch)
	}
	nb.backspace()
	if len(nb.digits) != 2 || string(nb.digits) != "12" {
		t.Errorf("Expected ['1','2'], got %v", nb.digits)
	}
}

func TestBackspace_Empty(t *testing.T) {
	nb := &NumericBuffer{}
	nb.backspace() // Should not panic
	if len(nb.digits) != 0 {
		t.Error("Expected empty after backspace on empty")
	}
}

func TestReset(t *testing.T) {
	nb := &NumericBuffer{mode: NumericHeight}
	for _, ch := range "123" {
		nb.append(ch)
	}
	nb.reset()
	if nb.mode != NumericNone || len(nb.digits) != 0 {
		t.Errorf("Expected reset state, got mode=%d digits=%v", nb.mode, nb.digits)
	}
}
```

---

## ANSI Renderer Test Template

```go
package main

import (
	"strings"
	"testing"
)

func TestRenderBox_Positioning_Normal(t *testing.T) {
	// 80x24 terminal, box should position at row=6, col=40
	ui := &uiRenderer{available: true, boxWidth: 40}
	// Render to buffer (need to refactor to return string instead of writing to file)
	// For now, just verify the positioning logic
	boxRow := 24 / 4  // = 6
	boxCol := 80 - 40 // = 40
	if boxRow != 6 || boxCol != 40 {
		t.Errorf("Expected row=6 col=40, got row=%d col=%d", boxRow, boxCol)
	}
}

func TestRenderBox_Positioning_Small(t *testing.T) {
	// Small terminal 40x12, box width 40
	boxRow := 12 / 4  // = 3
	boxCol := 40 - 40 // = 0 → clamped to 1
	if boxCol < 1 {
		boxCol = 1
	}
	if boxRow != 3 || boxCol != 1 {
		t.Errorf("Expected row=3 col=1, got row=%d col=%d", boxRow, boxCol)
	}
}

func TestRenderBox_ModeReal(t *testing.T) {
	// When useReal=true, modeStr should be "(real)"
	useReal := true
	var modeStr string
	if useReal {
		modeStr = "(real)"
	}
	if modeStr != "(real)" {
		t.Errorf("Expected '(real)', got %q", modeStr)
	}
}

func TestRenderBox_ModeFake(t *testing.T) {
	// When useReal=false and delta=0, modeStr should be "(fake)"
	useReal := false
	currentDelta := 0
	var modeStr string
	if useReal {
		modeStr = "(real)"
	} else if currentDelta != 0 {
		// delta case
	} else {
		modeStr = "(fake)"
	}
	if modeStr != "(fake)" {
		t.Errorf("Expected '(fake)', got %q", modeStr)
	}
}

func TestRenderBox_ModeDeltaPos(t *testing.T) {
	// When delta=+20, modeStr should be "(Δ+20)"
	useReal := false
	currentDelta := 20
	var modeStr string
	if useReal {
		modeStr = "(real)"
	} else if currentDelta != 0 {
		sign := "+"
		if currentDelta < 0 {
			sign = ""
		}
		modeStr = "(" + "Δ" + sign + "20" + ")"
	}
	if !strings.Contains(modeStr, "Δ") || !strings.Contains(modeStr, "+20") {
		t.Errorf("Expected delta mode string with 'Δ+20', got %q", modeStr)
	}
}

// ... additional rendering tests
```

---

## Integration Test Template

```go
package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestCommandMode_EnterViaCtrlBackslash tests entering command mode
func TestCommandMode_EnterViaCtrlBackslash(t *testing.T) {
	// This test requires running the long-term binary in a PTY
	// Pseudo-code:
	// 1. Start long-term binary with test command
	// 2. Send stdin: Ctrl+\ x3 within 500ms
	// 3. Verify /dev/tty contains ANSI codes for UI box
	// 4. Exit

	// For now, this is a placeholder showing structure
	t.Skip("Requires PTY test harness (TODO)")
}

// TestArrowUp_Increment tests UP arrow increments height
func TestArrowUp_Increment(t *testing.T) {
	// 1. Enter command mode
	// 2. Send UP arrow
	// 3. Verify height incremented by 1
	// 4. Send command to verify PTY height changed
	t.Skip("Requires PTY test harness (TODO)")
}

// ... additional integration tests
```

---

## Running Tests

```bash
# Run all tests with verbose output
go test ./... -v

# Run with race detection
go test ./... -race

# Run with coverage
go test ./... -cover

# Run specific test
go test -run TestKeyChar_Regular -v

# Run with timeout (tests shouldn't hang)
go test ./... -timeout 10s
```

---

## Next Steps

1. Copy templates into actual test files
2. Run `go test ./unit` to verify compilation
3. Fix any issues
4. Implement remaining tests from audit report
5. Run full suite with `-race` flag

