# Sprint: UI Rendering & Display
Generated: 2026-01-19 22:10:00
Confidence: HIGH
Status: READY FOR IMPLEMENTATION

## Sprint Goal
Implement /dev/tty-based UI rendering with ANSI escape codes for command mode display.

## Scope
**Deliverables:**
1. /dev/tty writer and ANSI escape code utilities
2. Modal UI box rendering (positioned 1/4 from top, right-aligned)
3. Dynamic content display (current settings, command help)

## Work Items

### P0: Terminal Output Manager
**Acceptance Criteria:**
- [ ] Open `/dev/tty` for writing (separate from stdin/stdout/stderr)
- [ ] Graceful fallback if /dev/tty unavailable (e.g., piped context)
- [ ] ANSI escape code helpers: clear line, move cursor, save/restore cursor, hide/show cursor
- [ ] Buffered writes to /dev/tty (atomic rendering of multi-line UI)
- [ ] Close /dev/tty on cleanup

**Technical Notes:**
- `tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)`
- ANSI codes:
  - Clear line: `\033[2K`
  - Move cursor: `\033[row;colH` (1-indexed)
  - Save/restore: `\033[s` / `\033[u`
  - Hide/show cursor: `\033[?25l` / `\033[?25h`
- Create helper functions: `clearLine()`, `moveCursor(row, col)`, `saveCursor()`, etc.
- Buffer writes in memory, flush once to avoid flicker

### P0: Modal Box Layout
**Acceptance Criteria:**
- [ ] Calculate position: row = termHeight/4, col = termWidth - boxWidth
- [ ] Box width ~40 columns
- [ ] Box height adapts to content (10-15 lines)
- [ ] Draw box border (Unicode box-drawing chars or ASCII +-|)
- [ ] Header line: "🖥  LONG-TERM ENABLED 🖥"
- [ ] Clear box area before rendering new content

**Technical Notes:**
- Get terminal size: `term.GetSize(int(os.Stdin.Fd()))`
- Box structure:
  ```
  ┌────────────────────────────────────┐
  │  🖥  LONG-TERM ENABLED 🖥           │
  ├────────────────────────────────────┤
  │ Term size: 80x100 (Δ+20)           │
  │                                    │
  │ UP/DOWN: ±1  Shift: ±20  Ctrl: ±200│
  │ n: set height  d: set delta        │
  │ space: toggle  r: reset  ESC: exit │
  └────────────────────────────────────┘
  ```
- Unicode box-drawing: `┌─┐│└┘├┤┬┴┼` (U+2500 series)
- Render each line with `moveCursor()` + content + `clearLine()`

### P0: Dynamic Content Display
**Acceptance Criteria:**
- [ ] Show current terminal size format:
  - Absolute mode: "Term size: WxH (fake)"
  - Delta mode: "Term size: WxH (Δ+N)" or "Term size: WxH (Δ-N)"
  - Real mode: "Term size: WxH (real)"
- [ ] Show toggle state indicator
- [ ] Show command shortcuts (arrow keys, n, d, space, r, ESC)
- [ ] Update display when settings change (height/delta adjusted)
- [ ] Show numeric input buffer when in n/d mode: "Enter height: 50_"

**Technical Notes:**
- Track state: current height, delta, mode (absolute/delta/real), toggle state
- Format strings for each line
- Refresh UI on: mode entry, key press (height change), toggle, numeric input
- Cursor positioning for numeric input prompt

## Dependencies
- Sprint 1 (Input Parsing) - needs mode state and key events

## Risks
**Terminal size too small for UI box**
- Mitigation: Detect if width < 50 or height < 20, show minimal UI or warning message

**Wrapped process output corruption**
- User selected "allow scroll/overlay" - UI needs refresh after wrapped output
- Mitigation: Refresh UI every 100ms while in command mode, or after detected output

**ANSI escape code compatibility**
- Mitigation: Test on multiple terminals, document limitations
