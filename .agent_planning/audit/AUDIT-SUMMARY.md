# Test Audit Summary: long-term Command-Mode

## Quick Reference

**Status**: ⚠️ **CRITICAL - 0% Test Coverage**

| Metric | Value |
|--------|-------|
| Automated tests | 0/86 (0%) |
| Complexity sources | 10 untested |
| P0 gaps (critical) | 8 |
| P1 gaps (important) | 12 |
| P2 gaps (nice-to-have) | 6 |
| Testability blockers | 3 |
| LOC | 795 |
| Goroutines | 5 concurrent |
| State machines | 3 |

---

## What Was Tested

The audit analyzed:

1. **Keyboard Parser** - Escape sequence state machine (26 untested scenarios)
2. **Numeric Buffer** - Input validation and accumulation (18 untested scenarios)
3. **ANSI Renderer** - UI rendering and positioning (12 untested scenarios)
4. **Mode State Machine** - Atomic operations and transitions (10 untested scenarios)
5. **Magic Detector** - Ctrl+\ timing and counter logic (8 untested scenarios)
6. **I/O Coordination** - Signal handling, PTY resizing, mode toggling
7. **Terminal Integration** - Raw mode, /dev/tty fallback, signal safety

---

## Critical Findings

### High-Risk Areas (P0)

All of these can cause silent failures:

1. **Keyboard parser state timeout** - Can leave parser stuck in ESC mode
2. **Race between mode toggle and PTY resize** - Can cause size mismatch internally
3. **Signal handler during I/O** - SIGWINCH handler can race with PTY.Close()
4. **Terminal raw mode cleanup** - Left in raw mode if crash occurs
5. **Numeric buffer overflow** - Range validation untested
6. **Magic detector edge case** - Counter logic on 4+ presses unclear
7. **ANSI rendering edge case** - Division by zero if terminal height=0
8. **/dev/tty fallback** - Behavior when unavailable untested

### Medium-Risk Areas (P1)

Common user flows that could break:

- Arrow key modifiers (Shift, Ctrl combinations)
- Backspace chains in numeric input
- ESC key to exit numeric mode
- Delta mode +/- prefix validation
- Height/delta toggle behavior
- Reset command synchronization
- Keyboard parser state leaks between events
- Terminal resize during command mode

---

## Test Implementation Roadmap

### Sprint 1: Keyboard & Input Parsing (2-3 days)

**26 unit tests** for escape sequence handling:
- Basic ASCII keys: a-z, 0-9, space (3 tests)
- Special keys: backspace, enter, ESC (3 tests)
- Arrow keys: up, down, left, right (4 tests)
- Arrow keys with modifiers: Shift, Ctrl, Shift+Ctrl (6 tests)
- Edge cases: timeout, state recovery, sequences (8 tests)
- Timing: exact window boundaries (2 tests)

**Catch**: Escape timeout bugs, state machine leaks

### Sprint 2: Numeric Input & UI (1-2 days)

**18 unit tests** for numeric buffer:
- Height validation: 1-9999 (5 tests)
- Delta validation: ±1 to ±9999 with sign (5 tests)
- Buffer operations: append, backspace, reset (3 tests)

**12 unit tests** for ANSI rendering:
- Positioning logic: normal, small, large terminals (4 tests)
- Mode indicators: real, fake, delta (4 tests)
- Numeric input display (2 tests)
- Box clearing (1 test)

**Catch**: Input validation bugs, display positioning issues

### Sprint 3: State Machines & Timing (1-2 days)

**10 unit tests** for mode state machine:
- Atomic operations: load, store, CAS patterns (5 tests)
- Height clamping: min/max bounds (3 tests)
- Delta bounds checking (2 tests)

**8 unit tests** for magic detector:
- 3-press detection (1 test)
- Window timeout (2 tests)
- Counter reset logic (3 tests)
- Mixed bytes filtering (2 tests)

**Catch**: Race conditions, timing edge cases

### Sprint 4: Integration Tests (2-3 days)

**12 integration tests** for full workflows:
- Enter command mode via Ctrl+\ (1 test)
- Exit via ESC (1 test)
- Arrow keys increment/decrement (2 tests)
- Numeric entry for height/delta (2 tests)
- Space toggle real/fake (1 test)
- Reset to defaults (1 test)
- PTY resize synchronization (2 tests)
- Mode toggle during SIGWINCH (1 test)
- Signal handler ordering (1 test)

**Catch**: Cross-component coordination bugs

### Sprint 5: CI Integration (1 day)

- Add `.github/workflows/test.yml`
- Run `go test ./... -race` on all PRs
- Generate coverage reports
- Add test status badge to README

---

## Testability Blockers & Solutions

### Blocker 1: PTY I/O Requires Real Terminal

**Problem**: `pty.StartWithSize()` needs actual TTY

**Solution**:
1. Extract PTY sizing into interface
2. Mock for unit tests
3. Use real PTY for integration tests only

### Blocker 2: SIGWINCH is OS-Level Signal

**Problem**: `signal.Notify()` can't be injected

**Solution**:
1. Extract signal handler into injectable function
2. Accept `chan os.Signal` parameter
3. Unit test coordination logic
4. Integration tests verify with real signals

### Blocker 3: Raw Mode Affects Global Terminal State

**Problem**: `term.MakeRaw()` modifies live terminal

**Solution**:
1. Run tests in isolated PTY environment
2. Or wrap `term` functions in interface
3. Mock for unit tests
4. Integration tests run in real terminal session

---

## Code Quality Issues Found

### Race Condition Risks

- **Atomic operations untested**: `atomic.Int32`, `atomic.Uint32`, `atomic.Bool`
- **Goroutine coordination**: 5 concurrent goroutines sharing atomics/channels
- **Signal handler timing**: SIGWINCH handler can race with PTY.Close()

**Mitigation**: Run `go test -race ./...` after implementation

### State Machine Issues

- **Keyboard parser**: 3 states, timeout logic, state transitions
- **Mode state**: 2 modes (Normal/Command), atomic updates
- **Numeric state**: 3 modes (None/Height/Delta)

**Mitigation**: Unit test each state machine independently

### Timing Bugs

- **Escape timeout**: 100ms window, off-by-one on comparison
- **Magic detector**: 500ms window, counter reset logic
- **SIGWINCH frequency**: May call handler multiple times

**Mitigation**: Test timing with explicit time.Sleep() and mocked time

---

## Effort Estimation

| Task | Days | Effort |
|------|------|--------|
| Sprint 1 (keyboard tests) | 2-3 | 40-60 lines per test × 26 = 1,040-1,560 LOC |
| Sprint 2 (numeric + UI tests) | 1-2 | 30-40 lines per test × 30 = 900-1,200 LOC |
| Sprint 3 (state machine tests) | 1-2 | 25-35 lines per test × 18 = 450-630 LOC |
| Sprint 4 (integration tests) | 2-3 | PTY harness + 12 tests = 800-1,200 LOC |
| Sprint 5 (CI + coverage) | 1 | GitHub Actions workflow + coverage reporting |
| **Total** | **8-11** | **~4,000-5,000 LOC of test code** |

**Expected outcome**: 85%+ code coverage, zero untested complexity sources

---

## Files Generated

```
.agent_planning/audit/
├── TEST-AUDIT-REPORT.md          (This audit, 450+ lines)
├── TEST-CASE-TEMPLATES.md        (Test code templates, 400+ lines)
└── AUDIT-SUMMARY.md              (This file)
```

All files saved to: `/Users/bmf/code/long-term/.agent_planning/audit/`

---

## Verification Checklist

Before starting implementation, verify:

- [ ] Read TEST-AUDIT-REPORT.md fully
- [ ] Understand P0/P1/P2 gaps (why each matters)
- [ ] Review testability blockers and proposed solutions
- [ ] Understand state machines (keyboard, mode, numeric)
- [ ] Review test templates for structure

Before merging tests:

- [ ] All 74 unit tests pass
- [ ] All 12 integration tests pass
- [ ] Coverage >= 80%
- [ ] `go test -race ./...` shows no race conditions
- [ ] No goroutine leaks (check with goroutine count at start/end)

---

## Recommended Next Steps

### Immediate (This Week)

1. Review this audit thoroughly
2. Set up Go test infrastructure (mkdir tests/)
3. Implement Sprint 1 (keyboard tests) - identify bugs early
4. Fix any bugs found in keyboard parser

### Short-term (Next 1-2 Weeks)

5. Implement Sprint 2 (numeric buffer + ANSI rendering)
6. Implement Sprint 3 (state machine + magic detector)
7. Create PTY test harness for integration tests

### Medium-term (Next 2-3 Weeks)

8. Implement Sprint 4 (integration tests)
9. Add CI configuration
10. Reach 85%+ coverage

### Ongoing

- Run tests on every PR
- Monitor coverage metrics
- Add tests for any bugs found

---

## Key Insights

1. **This is complex code**: 10 different complexity sources, 5 concurrent goroutines, 3 state machines

2. **No test failures today doesn't mean it's correct**: Manual testing creates false confidence; automated tests catch race conditions and edge cases

3. **Tests will reveal bugs**: Escape timeout logic, state leaks, and magic detector counter reset likely have subtle bugs

4. **High-value testing**: Each dollar spent on testing here prevents 10x debugging effort later

5. **Testability = better design**: Extracting PTY, signal, and terminal logic into interfaces will make code more maintainable

---

## Questions?

Refer to TEST-AUDIT-REPORT.md for:
- Detailed gap analysis (P0/P1/P2)
- Coverage matrix (complexity vs. tests)
- Quality assessment (red flags)
- Detailed recommendations

Refer to TEST-CASE-TEMPLATES.md for:
- Go test syntax and structure
- Example test implementations
- Running tests locally

