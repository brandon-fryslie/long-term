# Sprint: Command Integration & Polish
Generated: 2026-01-19 22:10:00
Confidence: HIGH
Status: READY FOR IMPLEMENTATION

## Sprint Goal
Wire up all commands, implement startup hint, handle edge cases and polish UX.

## Scope
**Deliverables:**
1. Arrow key commands (height adjustment)
2. Numeric entry commands (n/d)
3. Toggle and reset commands (space/r)
4. Startup hint display
5. UI refresh and SIGWINCH handling

## Work Items

### P0: Arrow Key Height Adjustment
**Acceptance Criteria:**
- [ ] UP/DOWN: adjust fake height by ±1
- [ ] Shift+UP/DOWN: adjust by ±20
- [ ] Ctrl+Shift+UP/DOWN: adjust by ±200
- [ ] Apply new height immediately (trigger SIGWINCH to resize PTY)
- [ ] Update UI to show new height value
- [ ] Clamp to valid range (1-9999)

**Technical Notes:**
- Parse key events from Sprint 1 keyboard parser
- Check modifiers: `if key.Code == UP && key.Shift && key.Ctrl { delta = 200 }`
- Update height/delta based on current mode (absolute vs delta)
- Send `sigwinch <- syscall.SIGWINCH` to trigger PTY resize
- Refresh UI after adjustment

### P0: Numeric Entry (n/d commands)
**Acceptance Criteria:**
- [ ] 'n' enters numeric mode for absolute height
- [ ] 'd' enters numeric mode for delta (show +/- prefix prompt)
- [ ] Display input buffer: "Enter height: 123_" or "Enter delta: +50_"
- [ ] Backspace removes last digit
- [ ] Enter validates and applies:
  - Height: 1-9999
  - Delta: +1 to +9999 or -1 to -9999
- [ ] ESC cancels and returns to command mode
- [ ] Show error message for invalid input

**Technical Notes:**
- Use numeric buffer from Sprint 1
- Validate: `height >= 1 && height <= 9999`
- For delta: require explicit +/- prefix, parse with `strconv.Atoi()`
- On validation failure: show error in UI for 2 seconds, stay in numeric mode
- On success: apply value, update UI, return to command mode

### P0: Toggle & Reset
**Acceptance Criteria:**
- [ ] Space: toggle between fake and real height (same as Ctrl+\×3 in normal mode)
- [ ] 'r': reset to default height/delta from command-line flags
- [ ] Update toggle state indicator in UI immediately
- [ ] Trigger SIGWINCH to apply new height

**Technical Notes:**
- Space: flip `useRealSize` atomic bool (line 141-142 pattern)
- Reset: restore to initial `fakeHeight` and `heightDelta` values from flags
- Send SIGWINCH after each change
- Refresh UI to show updated state

### P0: Startup Hint
**Acceptance Criteria:**
- [ ] Display one-line hint on program start: "long-term: Press Ctrl+\\ x3 for command mode"
- [ ] Write to stderr (not stdout, to avoid interfering with wrapped program)
- [ ] Only show if stderr is a terminal (skip if piped)
- [ ] Subtle formatting (gray text if terminal supports colors)

**Technical Notes:**
- Check: `term.IsTerminal(int(os.Stderr.Fd()))`
- ANSI gray: `\033[90m` (reset with `\033[0m`)
- Write before entering raw mode (line 214-219)
- Format: `fmt.Fprintf(os.Stderr, "\033[90mlong-term: Press Ctrl+\\ x3 for command mode\033[0m\n")`

### P0: UI Refresh & SIGWINCH Handling
**Acceptance Criteria:**
- [ ] Refresh UI every 100ms while in command mode (handle wrapped output scrolling)
- [ ] On SIGWINCH (terminal resize), recalculate UI position and re-render
- [ ] Update terminal size display when resize detected
- [ ] Ensure PTY size updated (existing SIGWINCH handler, lines 183-210)

**Technical Notes:**
- Add goroutine: `for commandMode { time.Sleep(100ms); refreshUI() }`
- In SIGWINCH handler (line 183-210), check if in command mode:
  - If yes: update UI position calc, re-render
  - Always: update PTY size (existing logic)
- Refresh renders entire UI (clear previous, draw new)

### P0: Clean Exit & Error Handling
**Acceptance Criteria:**
- [ ] ESC key exits command mode, clears UI, restores cursor
- [ ] Ctrl+C (SIGINT) exits program cleanly, restores terminal state
- [ ] Handle /dev/tty open failure gracefully (show error, disable command mode)
- [ ] Close /dev/tty on program exit

**Technical Notes:**
- ESC handling: clear UI area (write spaces + clear lines), restore cursor (ANSI `\033[u`)
- Add SIGINT handler: `signal.Notify(sigintC, syscall.SIGINT)`
  - On SIGINT: restore terminal (existing `defer term.Restore`, line 219), close PTY, exit
- /dev/tty error: `if err != nil { fmt.Fprintf(os.Stderr, "Warning: command mode unavailable\n") }`

## Dependencies
- Sprint 1 (Input Parsing) - keyboard events and mode state
- Sprint 2 (UI Rendering) - rendering functions

## Risks
**UI refresh conflicts with wrapped output**
- User selected "allow scroll/overlay" approach
- Mitigation: Refresh frequently (100ms), accept that UI may scroll off temporarily

**SIGWINCH during numeric input**
- Terminal resized while user typing height value
- Mitigation: Preserve numeric buffer state, update UI layout, continue input

**Startup hint timing**
- Hint may be quickly scrolled away by wrapped program output
- Mitigation: Acceptable as brief hint; users can discover via 'h' in command mode or docs
