# Definition of Done: Input Parsing & State Management

**Sprint**: SPRINT-20260119-input-parsing
**Confidence**: HIGH

## Acceptance Checklist

### Functional Requirements
- [ ] Keyboard parser correctly identifies single-char inputs (a-z, 0-9, space, ESC)
- [ ] Arrow keys parsed: UP, DOWN, LEFT, RIGHT
- [ ] Modified arrow keys parsed: Shift+UP, Ctrl+UP, Shift+Ctrl+UP (and DOWN variants)
- [ ] ESC key disambiguation works (standalone ESC vs ESC starting sequence)
- [ ] Key events emitted via channel to command handler
- [ ] Mode state machine: Normal ↔ Command transitions work
- [ ] Ctrl+\×3 enters command mode (existing magic detector still works)
- [ ] ESC exits command mode
- [ ] Input forwarding paused during command mode
- [ ] Numeric input buffer: 'n' starts height entry, 'd' starts delta entry
- [ ] Numeric buffer: digits accumulate, backspace deletes, Enter applies, ESC cancels
- [ ] Delta numeric entry requires +/- prefix

### Testing Requirements
- [ ] Manual test: Run `long-term -- bash`, press arrow keys, verify not sent to bash
- [ ] Manual test: Enter command mode, press ESC, verify exits cleanly
- [ ] Manual test: Enter 'n', type "100", press Enter, verify height changes (testable with tput lines)
- [ ] Manual test: Enter 'd', type "+20", press Enter, verify delta changes
- [ ] Manual test: Press ESC during numeric entry, verify canceled

### Code Quality
- [ ] No new linting errors
- [ ] Code follows existing patterns (magicDetector, atomic bool, channels)
- [ ] Keyboard parser documented (struct fields, state machine)
- [ ] Mode transitions documented

### Integration
- [ ] Works with existing PTY I/O proxying (lines 223-232)
- [ ] Works with existing SIGWINCH handler
- [ ] No race conditions (`go test -race` passes if tests exist)
- [ ] No goroutine leaks (all goroutines exit on program shutdown)

## Success Criteria

**This sprint is DONE when:**
1. All checkbox items above are checked
2. User can enter/exit command mode cleanly
3. Keyboard input is correctly parsed and routed based on mode
4. Numeric input for height/delta works end-to-end
5. No regressions in existing functionality (PTY wrapping still works)
