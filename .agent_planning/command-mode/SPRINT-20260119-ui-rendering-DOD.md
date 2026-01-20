# Definition of Done: UI Rendering & Display

**Sprint**: SPRINT-20260119-ui-rendering
**Confidence**: HIGH

## Acceptance Checklist

### Functional Requirements
- [ ] /dev/tty opens successfully for writing
- [ ] Graceful fallback if /dev/tty unavailable (error message shown)
- [ ] ANSI escape codes work: clear line, move cursor, save/restore, hide/show cursor
- [ ] Modal box renders at correct position (row = height/4, right-aligned)
- [ ] Box has border (Unicode box-drawing or ASCII)
- [ ] Header displays: "🖥  LONG-TERM ENABLED 🖥"
- [ ] Current settings display correctly:
  - Absolute mode: "Term size: WxH (fake)"
  - Delta mode: "Term size: WxH (Δ+N)" or "Term size: WxH (Δ-N)"
  - Real mode: "Term size: WxH (real)"
- [ ] Command shortcuts displayed in box
- [ ] Numeric input buffer shown during n/d entry: "Enter height: 123_"
- [ ] UI updates when settings change
- [ ] UI clears when exiting command mode

### Testing Requirements
- [ ] Visual test: UI appears at correct position in terminal
- [ ] Visual test: UI box is properly sized (~40 columns wide)
- [ ] Visual test: Terminal resize updates UI position correctly
- [ ] Visual test: Test on iTerm2 (macOS) - box renders correctly
- [ ] Visual test: Test on Terminal.app (macOS) - box renders correctly
- [ ] Visual test: Wrapped program output scrolls, UI overlays on top
- [ ] Edge case test: Terminal too small (width <50) - shows minimal UI or warning
- [ ] Edge case test: /dev/tty unavailable - shows error, disables command mode

### Code Quality
- [ ] No hardcoded escape sequences (use constants or helper functions)
- [ ] ANSI helpers documented: what each escape code does
- [ ] UI rendering code is in separate function(s) for clarity
- [ ] Buffer writes to /dev/tty (atomic rendering)

### Integration
- [ ] Works with Sprint 1 mode state (checks current mode before rendering)
- [ ] Works with Sprint 1 numeric buffer (displays input during n/d entry)
- [ ] UI doesn't corrupt wrapped program output
- [ ] Cursor restoration works (cursor in correct position after exiting command mode)

## Success Criteria

**This sprint is DONE when:**
1. All checkbox items above are checked
2. Command mode UI is visible and positioned correctly
3. UI dynamically updates to show current settings
4. Terminal resize is handled gracefully
5. UI clears cleanly on exit
6. Tested on at least 2 different terminal emulators successfully
