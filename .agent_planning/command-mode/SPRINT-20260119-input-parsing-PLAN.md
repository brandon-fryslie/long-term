# Sprint: Input Parsing & State Management
Generated: 2026-01-19 22:10:00
Confidence: HIGH
Status: READY FOR IMPLEMENTATION

## Sprint Goal
Extend input handling to parse keyboard events and implement command mode state machine.

## Scope
**Deliverables:**
1. Keyboard input parser that handles escape sequences
2. Command/Normal mode state machine
3. Numeric input buffer for n/d commands

## Work Items

### P0: Keyboard Input Parser
**Acceptance Criteria:**
- [ ] Parse single characters (a-z, 0-9, space, ESC, n, d, r)
- [ ] Parse arrow key escape sequences (UP=ESC[A, DOWN=ESC[B)
- [ ] Parse modified arrow keys (Shift+UP=ESC[1;2A, Ctrl+UP=ESC[1;5A, Shift+Ctrl+UP=ESC[1;6A)
- [ ] Handle ESC timeout disambiguation (50-100ms wait to detect if ESC starts sequence or is standalone)
- [ ] Emit parsed key events via channel

**Technical Notes:**
- Extend `magicDetector` pattern (lines 76-120 in main.go)
- Use `io.Writer` interface for `io.TeeReader` integration
- Create `KeyEvent` struct with `Code` (UP/DOWN/ESC/etc) and `Modifiers` (Shift/Ctrl flags)
- State machine: IDLE → ESC_RECEIVED (on 0x1B) → SEQUENCE_BUILDING → emit KeyEvent or timeout
- Buffer escape sequence bytes, parse on complete sequence

### P0: Mode State Machine
**Acceptance Criteria:**
- [ ] Track current mode (Normal vs Command) using atomic value or mutex
- [ ] Transition: Normal → Command on Ctrl+\×3 (keep existing magicDetector)
- [ ] Transition: Command → Normal on ESC key
- [ ] While in Command mode, intercept all keyboard input (don't forward to PTY)
- [ ] While in Normal mode, forward all input to PTY

**Technical Notes:**
- Add `type Mode int` with `const (ModeNormal Mode = iota; ModeCommand)`
- Use `atomic.Uint32` to store current mode (lock-free reads)
- Modify stdin→PTY goroutine (line 224-228) to check mode before `io.Copy`
- When in Command mode, send key events to command handler instead of PTY

### P0: Numeric Input Buffer
**Acceptance Criteria:**
- [ ] 'n' key enters numeric mode for absolute height
- [ ] 'd' key enters numeric mode for delta (requires +/- prefix)
- [ ] Accumulate digit keypresses (0-9)
- [ ] Backspace removes last digit
- [ ] Enter applies value and returns to command mode
- [ ] ESC cancels numeric input

**Technical Notes:**
- Add sub-state: Command mode → NumericEntry (n or d pressed)
- Buffer struct: `{mode: "n"|"d", digits: []rune, value: int}`
- Parse on Enter: validate range (1-9999 for n, ±1-9999 for d)
- Apply by updating height/delta and triggering SIGWINCH

## Dependencies
- None (foundational work)

## Risks
**Race condition in mode transitions**
- Mitigation: Use atomic operations for mode flag, channels for coordination

**Escape sequence timeout complexity**
- Mitigation: Test thoroughly on different terminals, adjust timeout if needed (50-100ms range)

**Terminal compatibility (different escape codes)**
- Mitigation: Test on iTerm2, Terminal.app, xterm; document any known issues
