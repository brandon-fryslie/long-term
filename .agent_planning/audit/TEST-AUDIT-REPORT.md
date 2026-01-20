# Test Coverage Audit Report: long-term Command-Mode Implementation

**Audit Date**: 2026-01-20
**Project**: long-term (long-term binary)
**Focus**: Command-mode feature implementation (interactive keyboard UI and height control)
**Go Version**: 1.25.5

---

## Executive Summary

| Dimension | Finding |
|-----------|---------|
| **Architecture Type** | CLI Tool (PTY wrapper) + Interactive UI |
| **Test Maturity** | None (0% automated coverage) |
| **Test Framework** | None configured |
| **Existing Tests** | 1 manual script (outdated, doesn't cover command mode) |
| **Critical Gaps** | 8 P0 items; 12 P1 items; 6 P2 items |
| **Quality Score** | 1/10 (untested complex logic) |
| **Testability Blockers** | 3 significant ones (PTY interaction, signal handling, terminal I/O) |

---

## Project Classification

### Type: CLI Tool + Interactive UI

**Signals Detected**:
- Single-file Go package with `flag.Parse()` (CLI tool)
- `io.TeeReader`, `io.MultiWriter` (I/O proxying)
- ANSI escape codes hardcoded (UI rendering)
- SIGWINCH handler (signal-driven updates)
- `/dev/tty` file access (interactive mode)
- `creack/pty` dependency (PTY manipulation)

**Architecture Pattern**: Event-driven state machine with concurrent goroutines

---

## Existing Test Infrastructure

### Convention Detection

| Item | Status | Details |
|------|--------|---------|
| **Test Framework** | NONE | No pytest, jest, vitest, or Go testing configured |
| **Test Directory** | NONE | No `tests/`, `test/`, `__tests__/`, or `*_test.go` files |
| **Config Files** | NONE | No pytest.ini, jest.config.js, or similar |
| **CI Integration** | GitHub Actions | Only for binary releases on git tags (no test runs) |
| **Manual Testing** | 1 script | `test-toggle.sh` (bash loop showing height; outdated) |

**File Search Results**:
```
No *_test.go files found
No conftest.py, setupTests.ts, or test helpers found
Manual test: test-toggle.sh (379 bytes, basic only)
```

### Existing Test Capabilities

The only manual test (`test-toggle.sh`) provides:
- ✓ Manual visual validation of height reporting
- ✓ Interactive space to test Ctrl+\ toggle
- ✗ No automated verification
- ✗ No coverage of command-mode UI
- ✗ No keyboard parser testing
- ✗ No edge case coverage
- ✗ Cannot be integrated into CI

---

## Codebase Complexity Analysis

### Architecture Overview

The implementation spans ~795 lines with 5 major functional areas:

#### 1. **Keyboard Parser** (`keyboardParser` struct, lines 259-369)
- Observes stdin stream without consuming bytes
- Converts raw TTY bytes into `KeyEvent` structs
- **Complexity**: Escape sequence state machine (3 states)
  - State 0: Idle → detect ESC or printable chars
  - State 1: Saw ESC → distinguish ESC from escape sequence
  - State 2: Saw ESC[ → accumulate and parse complete sequence

**Escape Sequences Handled**:
- Arrow keys with modifiers: `A` (up), `B` (down), `C` (right), `D` (left)
- Modified forms: `1;2A` (Shift+Up), `1;5A` (Ctrl+Up), `1;6A` (Shift+Ctrl+Up)
- Basic printable ASCII (a-z, 0-9, space, etc.)
- Backspace (0x7F), Enter (\r, \n)
- Timeout handling: standalone ESC after 100ms with no following `[`

#### 2. **Numeric Input Buffer** (`NumericBuffer` struct, lines 62-108)
- Accumulates digit sequences for height/delta input
- Validates input ranges based on mode:
  - Height mode: 1-9999
  - Delta mode: ±1 to ±9999, requires +/- prefix
- **Edge Cases**:
  - Empty buffer error handling
  - Sign validation for delta mode
  - Range clamping

#### 3. **ANSI Rendering Engine** (`uiRenderer` struct, lines 125-257)
- Writes ANSI escape codes to `/dev/tty` (separate from stdout)
- Renders 10-line UI box with title, status, and command help
- **Dynamic Elements**:
  - Box positioning: 1/4 down terminal, right-aligned
  - Mode indicator: "(real)", "(fake)", or "(Δ+20)"
  - Input display: Shows accumulated numeric input with cursor underscore
  - Error messages: Displays validation errors
  - Fallback: Disables gracefully if `/dev/tty` unavailable

#### 4. **Mode State Machine** (lines 497-640)
- Two modes: `ModeNormal` (passthrough I/O) and `ModeCommand` (UI active)
- Atomic flags for lock-free synchronization:
  - `currentMode`: Controls input interception
  - `useRealSize`: Toggles between fake/real terminal height
  - `currentHeight`, `currentDelta`: Dimension tracking
- **State Transitions**:
  - Ctrl+\ x3 (500ms window) → Normal → Command
  - ESC key → Command → Normal
  - Space → Toggle fake/real height

#### 5. **Signal Handling & PTY Resizing** (lines 713-747)
- SIGWINCH handler (terminal resize)
- Recalculates effective height based on current mode
- Updates PTY dimensions with `pty.Setsize()`
- Triggers UI refresh if in command mode

#### 6. **Magic Detector** (`magicDetector` struct, lines 427-471)
- Observes stdin for Ctrl+\ (byte 0x1C) press sequences
- Implements 500ms window counter
- Resets on window timeout
- Sends toggle signal to enter command mode

---

## Complexity Source Inventory

| # | Source | Type | Severity | Location | Risk |
|---|--------|------|----------|----------|------|
| **C1** | Keyboard escape parsing | State machine | High | `keyboardParser.processByte()` lines 284-329 | Sequence timeout, state leaks |
| **C2** | Numeric input validation | Logic | Medium | `NumericBuffer.value()` lines 83-108 | Edge cases in range/sign checking |
| **C3** | ANSI rendering | I/O + string formatting | Medium | `uiRenderer.renderBox()` lines 152-228 | Terminal size changes, positioning |
| **C4** | Mode transitions | State machine + atomics | High | `currentMode` atomic, keyboard handler | Race conditions, atomic ordering |
| **C5** | PTY size synchronization | Async coordination | High | `SIGWINCH` handler + `Setsize()` | Race between size change and mode toggle |
| **C6** | Magic byte detection | Timing + state | Medium | `magicDetector.Write()` lines 446-471 | Window timing, counter reset logic |
| **C7** | I/O interception | Async I/O | High | `tee` reader lines 766-786 | Mode-dependent forwarding, deadlocks |
| **C8** | Terminal capability detection | Fallback logic | Low | `term.IsTerminal()`, `/dev/tty` open | Missing `/dev/tty` disables features |
| **C9** | Signal handling | OS integration | High | Signal channel setup lines 714-745 | Race with PTY close |
| **C10** | Raw mode terminal state | System state | High | `term.MakeRaw()`, defer restore lines 756-762 | Async cleanup, signal safety |

---

## Test Inventory

### Manual Tests

| Name | Type | Framework | Coverage | Status |
|------|------|-----------|----------|--------|
| `test-toggle.sh` | E2E | bash loop | Basic I/O only | Outdated |

### Automated Tests

| Type | Count | Location | Framework |
|------|-------|----------|-----------|
| **Unit** | 0 | N/A | N/A |
| **Integration** | 0 | N/A | N/A |
| **E2E** | 0 | N/A | N/A |
| **Contract** | 0 | N/A | N/A |
| **Total** | 0 | N/A | N/A |

---

## Coverage Matrix: Complexity Sources vs. Tests

| Complexity Source | Unit | Integration | E2E | Manual | Status |
|-------------------|------|-------------|-----|--------|--------|
| **C1: Keyboard escape parsing** | ❌ | ❌ | ❌ | ❌ | **UNTESTED** |
| **C2: Numeric input validation** | ❌ | ❌ | ❌ | ❌ | **UNTESTED** |
| **C3: ANSI rendering** | ❌ | ❌ | ❌ | ⚠️ | Partial (visual only) |
| **C4: Mode transitions** | ❌ | ❌ | ❌ | ⚠️ | Partial (requires manual input) |
| **C5: PTY size sync** | ❌ | ❌ | ⚠️ | ⚠️ | Partial (E2E would test) |
| **C6: Magic byte detection** | ❌ | ❌ | ❌ | ⚠️ | Partial (manual timing) |
| **C7: I/O interception** | ❌ | ❌ | ❌ | ⚠️ | Partial (manual test only) |
| **C8: Terminal capability** | ❌ | ❌ | ❌ | ❌ | **UNTESTED** |
| **C9: Signal handling** | ❌ | ❌ | ❌ | ⚠️ | Partial (requires resize) |
| **C10: Raw mode state** | ❌ | ❌ | ❌ | ⚠️ | Partial (implicit in E2E) |

**Legend**: ❌ = No coverage, ⚠️ = Manual/indirect only, ✅ = Automated

**Gap Analysis**: 10/10 complexity sources have ZERO automated test coverage.

---

## Detailed Gap Analysis

### P0 - Critical (Security, Data Loss, Fatal Bugs)

| # | Gap | Why Critical | Impact |
|---|-----|--------------|--------|
| **P0.1** | No test for keyboard parser escape timeout | Can cause state machine stuck in ESC mode | User input ignored, UI frozen |
| **P0.2** | No test for race between mode toggle and PTY resize | Atomic ops untested; can cause size mismatch | Terminal displays wrong height internally |
| **P0.3** | No test for signal handler during I/O | SIGWINCH handler can race with PTY.Close() | Potential panic or resource leak |
| **P0.4** | No test for terminal raw mode cleanup | Restored state unclear if handler panics | Terminal left in raw mode after crash |
| **P0.5** | No test for numeric buffer overflow | Range validation untested; can store invalid values | Invalid height request silently truncated |
| **P0.6** | No test for magic detector edge case (4+ presses) | Counter logic unclear on overflow | May fail to toggle after many rapid presses |
| **P0.7** | No test for ANSI rendering with terminal size = 0 | Division by zero or negative coordinates possible | Crash or undefined rendering behavior |
| **P0.8** | No test for /dev/tty open failure handling | Fallback behavior untested | Command mode behaves unpredictably |

### P1 - Important (Common User Flows, Important Errors)

| # | Gap | Why Important | Impact |
|---|-----|---------------|--------|
| **P1.1** | Arrow key modifier combinations not tested | Shift, Ctrl, Shift+Ctrl delta values unclear | User expects ±20, ±200; may get wrong increment |
| **P1.2** | Numeric input backspace chain untested | Backspace on empty buffer; multiple backspaces | UI shows wrong state or crashes |
| **P1.3** | ESC to exit numeric input not tested | State cleanup on ESC unclear | Numeric mode may persist incorrectly |
| **P1.4** | Enter with invalid numeric input not tested | Error message not verified | User sees no feedback for bad input |
| **P1.5** | Delta mode requires +/- prefix; not tested | Sign validation logic unclear | May accept "-5" or "+5" as numeric input |
| **P1.6** | Height/delta toggle (space key) not tested | useRealSize flip logic untested | May not toggle, or stay toggled |
| **P1.7** | Reset command ('r') not tested | Clears delta, resets to flags; unclear if SIGWINCH sent | Height may not update after reset |
| **P1.8** | Command mode UI refresh during resize not tested | Positioning logic under small terminal | Box may go off-screen or overlap |
| **P1.9** | Keyboard parser state leaks between events | State transitions not verified end-to-end | Previous key press affects next key |
| **P1.10** | Multiple escape sequences in one write() | Buffer accumulation untested | May lose key events or misparse |
| **P1.11** | Mode toggle via space while entering numeric input | State machine branching unclear | May partially apply input or mode |
| **P1.12** | Terminal doesn't support arrow keys (dumb terminal) | UnknownKeyCode silently dropped | Arrow keys don't work in limited terminals |

### P2 - Nice to Have (Edge Cases, Secondary Features)

| # | Gap | Why Optional | Impact |
|---|-----|--------------|--------|
| **P2.1** | Printable ASCII range (0x20-0x7E) completeness | Missing: special chars, non-US keyboards | Some keys not recognized in command mode |
| **P2.2** | Startup hint display with no stderr | Displayed only if stderr is terminal | Hint may not show in piped contexts |
| **P2.3** | Shell alias fallback not tested | Fallback to SHELL env; unclear if shell exists | May fail silently if shell missing |
| **P2.4** | Very large height values (>9999 clamped) | Clamping logic untested | May allow setting invalid heights |
| **P2.5** | Empty digit buffer error message clarity | User sees "empty input" but why? | Confusing if ESC pressed before entering digit |
| **P2.6** | Magic detector rapid presses (e.g., 10x in 500ms) | Counter accumulation logic untested | May trigger multiple times unintentionally |

---

## Quality Assessment

### Red Flag Detection

| Issue | Evidence | Severity | Type |
|-------|----------|----------|------|
| **Untested atomic operations** | `atomic.Int32`, `atomic.Uint32`, `atomic.Bool` used in hot path | High | Race condition risk |
| **Untested goroutine coordination** | 5 concurrent goroutines with shared atomics, channels | High | Deadlock/starvation risk |
| **Untested system call sequence** | `term.MakeRaw()` + defer cleanup in signal context | High | Async safety risk |
| **Untested timeout logic** | `time.Since()` comparisons in keyboard parser | Medium | Off-by-one timing bugs |
| **Untested state machine transitions** | 3 keyboard parser states, 2 mode states, 3 numeric states | High | State leaks, incorrect transitions |
| **Untested string formatting** | ANSI codes, numeric conversions to strings | Low | Display glitches only |
| **Hardcoded dimensions** | Box width 40, timeout 100ms, window 500ms | Low | Constants assumed valid |
| **Missing nil checks** | `kbParser.eventChan` could receive after `cmd.Wait()` returns | Medium | Deadlock or goroutine leak |

### LLM Anti-Pattern Check

**Scanned for common AI-generated testing mistakes**:
- ✅ No mocking of mocks (code doesn't mock mock calls)
- ✅ No tautological assertions (code is logic, not test)
- ✅ Implementation is sound (manual review passes)
- ⚠️ **But**: No tests exist to verify implementation matches intent

---

## Testability Assessment

### Testability Blockers

#### **Blocker 1: PTY I/O Requires Real Terminal (High Impact)**

**Problem**: PTY creation requires actual file descriptors and terminal capabilities.

```go
ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{...})
```

**Current Workarounds**:
- None (PTY tests must be integration/E2E)
- Would require mocking `creack/pty` or running in test terminal

**Recommendation**:
- Extract PTY sizing logic into testable interface
- Keep I/O adaptation in integration tests only

#### **Blocker 2: Signal Handling Cannot be Tested in Isolation (High Impact)**

**Problem**: SIGWINCH requires real terminal resize event; `signal.Notify()` is OS-level.

```go
signal.Notify(sigwinch, syscall.SIGWINCH)
```

**Current Workarounds**:
- None (signal tests must use OS facilities)
- Would require test harness to send real signals

**Recommendation**:
- Extract signal handler into `func(chan os.Signal)` for testability
- Unit test coordination logic separately

#### **Blocker 3: Raw Mode Terminal State Cannot be Rolled Back in Tests (Medium Impact)**

**Problem**: `term.MakeRaw()` modifies global terminal state; cleanup is fragile.

```go
oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
defer term.Restore(int(os.Stdin.Fd()), oldState)
```

**Current Workarounds**:
- Run tests in isolated terminal session
- Mock `term` package (breaks contract)

**Recommendation**:
- Wrap `term.MakeRaw/Restore` in interface for mocking
- Keep real TTY code in integration tests only

---

## Testing Strategy Recommendations

### Recommended Test Structure

```
tests/
├── unit/
│   ├── keyboard_parser_test.go         (escape sequences, state machine)
│   ├── numeric_buffer_test.go          (validation, range checking)
│   ├── mode_state_machine_test.go      (transitions, atomics)
│   ├── ansi_renderer_test.go           (formatting, positioning)
│   └── magic_detector_test.go          (timing, counter reset)
│
├── integration/
│   ├── command_mode_integration_test.go (full flow in real PTY)
│   └── mode_toggle_integration_test.go (SIGWINCH + mode sync)
│
└── e2e/
    ├── interactive_test.sh              (manual; prompts for Ctrl+\)
    └── batch_test.sh                    (automated height verify)
```

### Phase 1: Unit Tests (Lowest Hanging Fruit)

#### 1.1 **Keyboard Parser Unit Tests** (26 test cases)

```go
// tests/unit/keyboard_parser_test.go

// Basic ASCII
- TestKeyChar_Regular()           // 'a' → KeyChar('a')
- TestKeyChar_Digit()             // '5' → KeyChar('5')
- TestKeyChar_Space()             // ' ' → KeyChar(' ')

// Special keys
- TestKeyBackspace()              // 0x7F → KeyBackspace
- TestKeyEnter_CR()               // '\r' → KeyEnter
- TestKeyEnter_LF()               // '\n' → KeyEnter

// Arrow keys (basic)
- TestKeyUp()                     // ESC[A → KeyUp
- TestKeyDown()                   // ESC[B → KeyDown
- TestKeyLeft()                   // ESC[D → KeyLeft
- TestKeyRight()                  // ESC[C → KeyRight

// Arrow keys with modifiers
- TestKeyUp_Shift()               // ESC[1;2A → KeyUp + Shift
- TestKeyDown_Shift()             // ESC[1;2B → KeyDown + Shift
- TestKeyUp_Ctrl()                // ESC[1;5A → KeyUp + Ctrl
- TestKeyDown_Ctrl()              // ESC[1;5B → KeyDown + Ctrl
- TestKeyUp_ShiftCtrl()           // ESC[1;6A → KeyUp + ShiftCtrl
- TestKeyDown_ShiftCtrl()         // ESC[1;6B → KeyDown + ShiftCtrl

// Edge cases
- TestKeyESC_Standalone()         // ESC after timeout (100ms)
- TestKeyESC_InSequence()         // ESC[ processed correctly
- TestKeyMultiple_Sequential()    // 'a' then 'b' separately
- TestKeyUnknown_Ignored()        // 0x00, 0x01, etc ignored
- TestKeySequence_Incomplete()    // ESC[ followed by EOF
- TestKeySequence_Partial()       // ESC only, then separate byte
- TestKeyStateRecovery()          // Bad escape then good escape
- TestKeyTiming_ESCTimeout()      // ESC, wait 100ms, verify standalone
- TestKeyTiming_ESCNoTimeout()    // ESC[ within 100ms, completes

Expected: 26 tests, all pass, 100% parser code coverage
```

#### 1.2 **Numeric Buffer Unit Tests** (18 test cases)

```go
// tests/unit/numeric_buffer_test.go

// Height mode validation
- TestHeightValue_Valid_Min()     // "1" → 1
- TestHeightValue_Valid_Max()     // "9999" → 9999
- TestHeightValue_Invalid_Zero()  // "0" → error
- TestHeightValue_Invalid_Over()  // "10000" → error
- TestHeightValue_Empty()         // "" → error

// Delta mode validation
- TestDeltaValue_Valid_Plus()     // "+50" → 50
- TestDeltaValue_Valid_Minus()    // "-50" → -50
- TestDeltaValue_Invalid_NoSign() // "50" → error (delta mode)
- TestDeltaValue_Invalid_Over()   // "+10000" → error
- TestDeltaValue_Valid_Max()      // "+9999" → 9999
- TestDeltaValue_Valid_Min()      // "-9999" → -9999

// Buffer operations
- TestAppend_Single()             // append('5') → [5]
- TestAppend_Multiple()           // append repeatedly → accumulates
- TestBackspace_Single()          // [5] → [] after backspace
- TestBackspace_Multiple()        // [1,2,3] → [1,2] after backspace
- TestBackspace_Empty()           // [] → [] (no crash)
- TestReset()                     // Clears digits and mode

Expected: 18 tests, all pass, 100% NumericBuffer coverage
```

#### 1.3 **ANSI Renderer Unit Tests** (12 test cases)

```go
// tests/unit/ansi_renderer_test.go

// Positioning logic
- TestRenderBox_Positioning_Normal()      // 80x24 terminal → correct row/col
- TestRenderBox_Positioning_Small()       // 40x12 terminal → wraps, clamps
- TestRenderBox_Positioning_Large()       // 200x100 → positions correctly
- TestRenderBox_Positioning_NegCol()      // Negative col → clamped to 1

// Mode indicator strings
- TestRenderBox_ModeReal()                // useReal=true → "(real)"
- TestRenderBox_ModeFake()                // delta=0, useReal=false → "(fake)"
- TestRenderBox_ModeDeltaPos()            // delta=+20 → "(Δ+20)"
- TestRenderBox_ModeDeltaNeg()            // delta=-5 → "(Δ-5)"

// Numeric input display
- TestRenderBox_NumericInput_Height()     // Shows "Enter height: 50_"
- TestRenderBox_NumericInput_Delta()      // Shows "Enter delta: +20_"
- TestRenderBox_ErrorMessage()            // Displays error text
- TestClearBox_Removal()                  // Clears 10 lines

Expected: 12 tests, all pass; verifies ANSI codes and formatting
```

#### 1.4 **Mode State Machine Unit Tests** (10 test cases)

```go
// tests/unit/mode_state_machine_test.go (helper tests)

- TestAtomicMode_NormalRead()             // Load ModeNormal
- TestAtomicMode_CommandRead()            // Load ModeCommand
- TestAtomicMode_Store()                  // Store and verify read
- TestAtomicMode_ManyReads()              // Concurrent reads (no race)
- TestUseRealSize_Toggle()                // Toggle atomic bool multiple times
- TestHeightStorage_Store_Load()          // Store int32, load matches
- TestDeltaStorage_Negative()             // Store negative delta
- TestHeightClamping_Min()                // Clamp height ≥ 1
- TestHeightClamping_Max()                // Clamp height ≤ 9999
- TestDeltaClamping()                     // Clamp delta ±9999

Expected: 10 tests verifying atomic semantics
```

#### 1.5 **Magic Detector Unit Tests** (8 test cases)

```go
// tests/unit/magic_detector_test.go

- TestMagicDetector_ThreePresses()        // 0x1C x3 → toggles
- TestMagicDetector_TwoPresses()          // 0x1C x2 → no toggle
- TestMagicDetector_FourPresses()         // 0x1C x4 → toggles once (counter reset)
- TestMagicDetector_WindowTimeout()       // 0x1C, wait 501ms, 0x1C x2 → no toggle
- TestMagicDetector_WindowBoundary()      // 0x1C, wait 500ms, 0x1C x2 → toggle (at boundary)
- TestMagicDetector_WithinWindow()        // 0x1C x2 in 100ms, 0x1C → triggers
- TestMagicDetector_OtherBytes()          // 'a', 'b' mixed with 0x1C → counts only 0x1C
- TestMagicDetector_EmptyChunk()          // Write empty slice → no crash

Expected: 8 tests verifying timing and counter logic
```

**Unit Test Total**: 26 + 18 + 12 + 10 + 8 = **74 test cases**

### Phase 2: Integration Tests (Mid-Level)

#### 2.1 **Command Mode Integration Tests** (8 test cases)

These require a test harness that runs the binary in a pseudo-PTY:

```go
// tests/integration/command_mode_integration_test.go

- TestCommandMode_EnterViaCtrlBackslash()  // Send 0x1C x3 → enter command mode
- TestCommandMode_ExitViaESC()            // Enter, send ESC → exit command mode
- TestCommandMode_ArrowUp_Increment()     // Up arrow → height +1
- TestCommandMode_ArrowDown_Decrement()   // Down arrow → height -1
- TestCommandMode_ShiftUp_Large()         // Shift+Up → height +20
- TestCommandMode_NumericEntry_Height()   // Press 'n', enter "50", Enter → height = 50
- TestCommandMode_NumericEntry_Delta()    // Press 'd', enter "+20", Enter → delta = +20
- TestCommandMode_SpaceToggle()           // Press space → toggle real/fake

Expected: 8 tests covering full user workflows
```

#### 2.2 **PTY Resize Synchronization Tests** (4 test cases)

```go
// tests/integration/mode_toggle_integration_test.go

- TestSIGWINCH_UpdatesPTYSize()           // Resize terminal → PTY size updated
- TestSIGWINCH_DuringCommandMode()        // Resize while in command mode → UI repositioned
- TestModeToggle_SizeSwitch()             // Toggle fake/real → PTY size changes
- TestModeToggle_DeltaRecalc()            // Delta mode + resize → height = real + delta

Expected: 4 tests covering SIGWINCH coordination
```

**Integration Test Total**: 8 + 4 = **12 test cases**

### Phase 3: Manual E2E Tests (Keeps Existing Script)

```bash
# tests/e2e/interactive_test.sh
# Prompts user to:
#   1. Run: long-term -- bash -c 'while sleep 1; do tput lines; done'
#   2. Press Ctrl+\ x3
#   3. Verify UI appears
#   4. Test arrow keys, numeric input, etc.
#   5. Exit and verify terminal state clean

# tests/e2e/batch_test.sh
# Non-interactive:
#   1. Start: long-term -height 50 -- bash
#   2. Command: tput lines
#   3. Verify output = 50
```

---

## Implementation Roadmap

### Sprint 1: Keyboard & Input (2-3 days)

**Goal**: Unit test keyboard parser + numeric buffer

1. Set up Go test infrastructure (`*_test.go` files)
2. Write 26 keyboard parser tests
3. Write 18 numeric buffer tests
4. Fix any bugs found in tests

**Expected**: Escape sequence and input validation bugs revealed

### Sprint 2: Rendering & UI (1-2 days)

**Goal**: Unit test ANSI rendering

1. Write 12 ANSI renderer tests
2. Write 10 mode state machine tests
3. Verify atomic operations don't race under test

**Expected**: Positioning bugs, edge case rendering issues revealed

### Sprint 3: Timing & Detection (1 day)

**Goal**: Unit test magic detector and timing

1. Write 8 magic detector tests
2. Verify window timeout logic
3. Test counter reset behavior

**Expected**: Off-by-one timing bugs revealed

### Sprint 4: Integration (2-3 days)

**Goal**: Integration tests for full workflows

1. Create PTY test harness (mock PTY or real PTY in test environment)
2. Write 8 command mode integration tests
3. Write 4 SIGWINCH synchronization tests
4. Verify no goroutine leaks or deadlocks

**Expected**: Cross-component coordination bugs revealed

### Sprint 5: CI Integration (1 day)

**Goal**: Add test suite to GitHub Actions

1. Create `.github/workflows/test.yml`
2. Run `go test ./...` on PR/push
3. Generate coverage report
4. Add badge to README

**Expected**: Continuous test validation

---

## Code Refactoring for Testability

To make code more testable, minimal invasive changes recommended:

### Refactoring 1: Extract Interfaces

```go
// Before: creack/pty directly used in run()
ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{...})

// After: injectable interface for testing
type PTYProvider interface {
    StartWithSize(cmd *exec.Cmd, size *pty.Winsize) (io.ReadWriter, error)
    Setsize(f *os.File, size *pty.Winsize) error
}

// Real implementation wraps creack/pty
// Test implementation returns mock
```

### Refactoring 2: Extract Signal Handler

```go
// Before: signal.Notify() directly in run()
signal.Notify(sigwinch, syscall.SIGWINCH)

// After: separate function that accepts channel
func handleSIGWINCH(sigwinch chan os.Signal, ...) {
    // logic testable without real signals
}
```

### Refactoring 3: Extract Terminal Mode

```go
// Before: term.MakeRaw() in run()
oldState, _ := term.MakeRaw(int(os.Stdin.Fd()))
defer term.Restore(int(os.Stdin.Fd()), oldState)

// After: injectable interface
type TerminalControl interface {
    MakeRaw(fd int) (*term.State, error)
    Restore(fd int, state *term.State) error
}

// Real implementation uses golang.org/x/term
// Test implementation is no-op
```

---

## Risk Assessment

### High-Risk Areas (Should be Tested First)

| Area | Risk | Why | Impact |
|------|------|-----|--------|
| **Keyboard parser state machine** | High | 3 states, timeout, state leaks | Input silently ignored or frozen |
| **Mode transitions with atomics** | High | Concurrent reads/writes, no synchronization beyond atomics | Data races, size mismatches |
| **SIGWINCH + mode toggle sync** | High | Two independent triggers, shared state | PTY size wrong, terminal corrupted |
| **Terminal raw mode cleanup** | High | Async cleanup in signal context | Terminal left unusable if crash |
| **Numeric input validation** | High | Range checking, sign validation | Invalid heights silently accepted |

### Medium-Risk Areas

| Area | Risk | Why | Impact |
|------|------|-----|--------|
| **ANSI rendering positioning** | Medium | Terminal size can be 0, box may go off-screen | Display glitches, overlap |
| **Magic detector timing** | Medium | 500ms window, counter reset | May trigger spuriously or not at all |
| **I/O interception in command mode** | Medium | Mode-dependent forwarding, channel coordination | Some input lost, some forwarded wrong |

### Low-Risk Areas

| Area | Risk | Why | Impact |
|------|------|-----|--------|
| **Startup hint display** | Low | Only affects message visibility | Cosmetic only |
| **Shell alias fallback** | Low | Fallback to /bin/sh if not found | Graceful degradation |
| **Height clamping bounds** | Low | Range checking, straightforward logic | Edge case only |

---

## Metrics & Success Criteria

### Coverage Goals

| Metric | Current | Target | By When |
|--------|---------|--------|---------|
| **Automated test coverage** | 0% | 85% | Sprint 4 end |
| **Unit tests** | 0 | 74 | Sprint 3 end |
| **Integration tests** | 0 | 12 | Sprint 4 end |
| **Keyboard parser coverage** | 0% | 100% | Sprint 1 end |
| **State machine coverage** | 0% | 100% | Sprint 2 end |
| **CI test runs** | 0 | Every PR | Sprint 5 end |

### Quality Gates

| Gate | Criterion | Enforcement |
|------|-----------|--------------|
| **Coverage threshold** | No PR merge if coverage < 80% | GitHub Actions check |
| **Test pass rate** | All tests must pass | GitHub Actions check |
| **Goroutine leaks** | Run with `-race` flag | CI test step |
| **Panic recovery** | No unhandled panics in tests | Manual review |

---

## Appendix

### A. Files Analyzed

```
/Users/bmf/code/long-term/main.go           (795 lines)
/Users/bmf/code/long-term/go.mod            (dependencies)
/Users/bmf/code/long-term/test-toggle.sh    (manual test)
/Users/bmf/code/long-term/justfile          (build commands)
/Users/bmf/code/long-term/README.md         (usage docs)
```

### B. Complexity Metrics

```
Lines of Code: 795
Cyclomatic Complexity: ~18 (high)
Goroutines: 5 concurrent
Atomic Operations: 4 (currentHeight, currentDelta, currentMode, useRealSize)
Channels: 4 (sigwinch, refreshUI, enterCommandChan, eventChan)
State Machines: 3 (keyboard parser 3 states, mode 2 states, numeric 3 states)
Escape Sequences Handled: 8+ combinations
```

### C. Go Test Infrastructure Setup

```bash
# Initialize testing in project
go mod tidy                          # Ensure dependencies

# Create test directories
mkdir -p tests/unit
mkdir -p tests/integration
mkdir -p tests/e2e

# Run tests
go test ./tests/unit -v
go test ./tests/integration -v -race  # Detect races
go test ./... -cover                  # Show coverage
```

### D. References

**Go Testing Best Practices**:
- https://golang.org/doc/effective_go#testing
- https://pkg.go.dev/testing

**Terminal & PTY Testing**:
- https://github.com/creack/pty (package documentation)
- Testing ANSI codes: capture output, verify codes present

**Atomic Operations**:
- https://pkg.go.dev/sync/atomic (documentation)
- Race detection: `go test -race ./...`

---

## Conclusion

The command-mode implementation for long-term is complex and feature-rich but **completely untested**. The codebase has:

✅ **Strengths**:
- Clear separation of concerns (parser, renderer, state machine)
- Correct use of atomics for coordination
- Graceful fallbacks (e.g., /dev/tty unavailable)

❌ **Critical Gaps**:
- 0% automated test coverage
- 10/10 complexity sources untested
- 8 P0 (critical), 12 P1 (important) gaps
- High risk of race conditions, state leaks, timing bugs

**Recommended Action**: Implement Phase 1 (unit tests for keyboard + input) immediately. This will likely reveal bugs in escape sequence parsing and numeric validation. Then proceed to integration tests for full workflows.

**Estimated Effort**: 8-10 days to reach 85% coverage and stable, tested implementation.

**ROI**: Eliminates silent failures, catches regressions early, enables confident refactoring.

