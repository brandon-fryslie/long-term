# Definition of Done: Command Integration & Polish

**Sprint**: SPRINT-20260119-integration
**Confidence**: HIGH

## Acceptance Checklist

### Functional Requirements
- [ ] Arrow keys adjust height:
  - UP/DOWN: ±1
  - Shift+UP/DOWN: ±20
  - Shift+Ctrl+UP/DOWN: ±200
- [ ] Height adjustment triggers PTY resize immediately
- [ ] Height clamped to valid range (1-9999)
- [ ] 'n' command: enter height, validate, apply
- [ ] 'd' command: enter delta (+/-1 to +/-9999), validate, apply
- [ ] Invalid numeric input shows error message
- [ ] Space toggles between fake and real height
- [ ] 'r' resets to default (original command-line flags)
- [ ] Startup hint displays: "long-term: Press Ctrl+\\ x3 for command mode"
- [ ] Startup hint only shows if stderr is a terminal
- [ ] UI refreshes every 100ms while in command mode
- [ ] SIGWINCH during command mode: UI repositioned and re-rendered
- [ ] ESC exits command mode cleanly (UI cleared, cursor restored)
- [ ] Ctrl+C exits program cleanly (terminal state restored)
- [ ] /dev/tty open failure shows warning, disables command mode

### Testing Requirements
- [ ] End-to-end test: Enter command mode, press UP 5 times, verify height increased by 5 (using `tput lines`)
- [ ] End-to-end test: Enter command mode, Shift+UP, verify height increased by 20
- [ ] End-to-end test: Enter command mode, 'n', type "500", Enter, verify height = 500
- [ ] End-to-end test: Enter command mode, 'd', type "+50", Enter, verify delta = +50
- [ ] End-to-end test: Enter command mode, space, verify toggle to real height
- [ ] End-to-end test: Enter command mode, 'r', verify reset to original flags
- [ ] End-to-end test: Run wrapped program that outputs continuously (e.g., `yes`), enter command mode, verify UI visible
- [ ] End-to-end test: Resize terminal while in command mode, verify UI repositions
- [ ] End-to-end test: Press Ctrl+C, verify terminal restored to normal mode
- [ ] End-to-end test: Run with piped stderr (`long-term ... 2>/dev/null`), verify no startup hint shown
- [ ] Integration test: Run with vim, enter command mode, adjust height, exit, verify vim still functional
- [ ] Integration test: Run with tmux, enter command mode, verify no conflicts
- [ ] Edge case test: Enter 'n', type "99999" (too large), verify error message
- [ ] Edge case test: Enter 'd', type "50" (no +/-), verify error message

### Code Quality
- [ ] No new linting errors
- [ ] All goroutines properly shut down on exit
- [ ] No race conditions (`go test -race` clean if tests exist)
- [ ] Code documented: command handlers, refresh logic, SIGWINCH integration
- [ ] Error messages user-friendly

### Integration
- [ ] All Sprint 1 deliverables integrated (keyboard parser, mode state, numeric buffer)
- [ ] All Sprint 2 deliverables integrated (UI rendering, /dev/tty, ANSI codes)
- [ ] No regressions in existing functionality:
  - [ ] PTY wrapping works
  - [ ] Simple toggle (Ctrl+\×3 in normal mode) still works
  - [ ] SIGWINCH handling for wrapped process works
  - [ ] Width pass-through works
  - [ ] Shell fallback for aliases works

### Documentation
- [ ] README.md updated with command mode usage
- [ ] test-toggle.sh updated or new test script created
- [ ] Command shortcuts documented in UI and/or help screen

## Success Criteria

**This sprint is DONE when:**
1. All checkbox items above are checked
2. Full command mode workflow works end-to-end
3. All commands (arrows, n, d, space, r, ESC) function correctly
4. Startup hint is visible and helpful
5. UI refresh handles wrapped output gracefully
6. Terminal resize is handled smoothly
7. Clean exit and error handling verified
8. Tested with at least 3 different wrapped programs (bash, vim, long-running process)
9. Zero known critical bugs or regressions
