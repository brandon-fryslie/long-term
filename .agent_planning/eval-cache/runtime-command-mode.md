# Runtime Behavior: Command Mode

**Scope**: command-mode implementation in long-term
**Last Updated**: 2026-01-20
**Confidence**: FRESH

## Known Runtime Behaviors

### Critical Bug: Delta Default Value
**Status**: BROKEN
**Location**: main.go:373
**Symptom**: `-height N` flag doesn't work; always reports (real_height + 2000)

**Root Cause**:
```go
heightDelta := flag.Int("delta", 2000, "...")  // WRONG: should be 0
```

**Why This Breaks Everything**:
- Default delta=2000 means "delta mode" is always active
- SIGWINCH handler checks `if delta != 0` (line 723)
- Even when user specifies `-height 50`, delta=2000 overrides it
- Result: Reports 2024 instead of 50 (assuming real height is 24)

**Impact Scope**:
- Primary use case (fixed height) completely broken
- Affects initial PTY setup (line 489-495)
- Affects SIGWINCH handler (line 722-730)
- Affects reset command 'r' (line 650-651)

**Verified Broken**:
```bash
$ ./long-term -height 50 -- bash -c 'tput lines'
2024  # WRONG! Should be 50
```

**Verified Working (Workaround)**:
```bash
$ ./long-term -height 50 -delta 0 -- bash -c 'tput lines'
50  # Correct when delta explicitly set to 0
```

**Fix Required**: Change default from 2000 to 0

---

### Working: Delta Mode
**Status**: WORKING
**Location**: main.go:489-495, 722-730

**Verified Behavior**:
```bash
$ ./long-term -delta +20 -- bash -c 'tput lines'
44  # Correct (24 real + 20 delta = 44)
```

**Data Flow**:
1. User passes `-delta +20`
2. initialDelta = 20
3. effectiveHeight = realHeight + delta = 24 + 20 = 44
4. PTY initialized with 44 rows
5. SIGWINCH handler: `if delta != 0` → uses delta mode
6. Result: Correctly reports 44

---

### Warning: /dev/tty Unavailable
**Status**: EXPECTED in non-interactive contexts
**Location**: main.go:133-136, 513-515

**Behavior**:
- Command mode requires `/dev/tty` for UI overlay
- Non-interactive contexts (piped, scripted) don't have /dev/tty
- Warning appears: "Warning: /dev/tty unavailable, command mode UI disabled"
- Program continues but command mode is disabled

**Not a Bug**: Expected limitation of the design

---

## Cannot Verify Without Interactive Terminal

The following behaviors require real terminal interaction:
- Command mode UI rendering
- Keyboard input parsing (Ctrl+\×3, arrow keys, etc.)
- Magic key sequence detection
- UI refresh timing (100ms)
- SIGWINCH during command mode
- Cursor save/restore

**Reason**: Requires /dev/tty, terminal raw mode, and real keyboard input

---

## Architectural Patterns (Verified Correct)

### Atomic State Management
- `currentHeight` atomic.Int32 - single source of truth
- `currentDelta` atomic.Int32 - single source of truth  
- `currentMode` atomic.Uint32 - single source of truth
- `useRealSize` atomic.Bool - single source of truth

**Pattern**: Lock-free access from multiple goroutines

### Single Enforcer Pattern
- SIGWINCH handler (line 715-745): sole enforcer of PTY resize
- Command handler (line 573-687): sole processor of keyboard events
- magicDetector (line 427-471): sole detector of Ctrl+\ sequence

**Pattern**: Each invariant has exactly one enforcement point

### Data Flow (When Working)
```
User Input → Flag Parsing → Atomics → PTY Setup → SIGWINCH → PTY Resize
```

**When Broken** (current state with delta=2000 default):
```
-height 50 → height=50, delta=2000 → Uses delta mode → Reports 2024
```

**When Fixed** (delta=0 default):
```
-height 50 → height=50, delta=0 → Uses height mode → Reports 50
```

---

## Test Results

### Automated Tests Available
- `go vet`: PASS
- `just build`: PASS

### Manual Tests Performed
| Test | Expected | Actual | Status |
|------|----------|--------|--------|
| `-height 50` | 50 | 2024 | ❌ BROKEN |
| `-delta +20` | real+20 | 44 (24+20) | ✅ WORKING |
| `-height 50 -delta 0` | 50 | 50 | ✅ WORKING (workaround) |

### Tests Blocked (Require Interactive Terminal)
- All command mode functionality (63 DoD criteria)
- See WORK-EVALUATION-command-mode-20260120.md for full list

---

## Reuse Guidance

**Use this cache when**:
- Evaluating fixes to the delta default bug
- Testing PTY height/delta functionality
- Verifying SIGWINCH behavior

**Don't trust for**:
- Interactive command mode features (not verified)
- Edge cases with vim, tmux, etc. (not tested)
- Numeric input validation (not tested interactively)

**Re-evaluation Needed If**:
- main.go lines 373, 489-495, or 722-730 change
- Flag parsing logic changes
- SIGWINCH handler changes

---

## Additional Bugs Found (2026-01-20 03:15)

### Bug: Height Validation Missing (Initial Flags)
**Status**: BROKEN
**Location**: main.go:373, no validation after flag parsing
**Severity**: HIGH

**Test Evidence**:
```bash
$ ./long-term -height 99999 -delta 0 -- bash -c 'tput lines'
-1  # Invalid PTY, tput reports error
```

**Root Cause**: Flags accepted without validation
- NumericBuffer validates (lines 94-96): 1-9999 range
- But initial flags never validated
- PTY setup uses invalid values → broken terminal

**Fix**: Add validation after flag.Parse(), before run()

---

### Bug: Delta Unbounded in Arrow Key Adjustments
**Status**: BROKEN
**Location**: main.go:672-673
**Severity**: MEDIUM

**Code**:
```go
if currentDelta.Load() != 0 {
    currentDelta.Add(int32(delta))  // NO CLAMPING
}
```

**Issue**: Arrow keys in delta mode have no bounds
- Height mode clamps (lines 676-680): 1-9999
- Delta mode doesn't clamp at all
- User can press Shift+Ctrl+UP many times → delta=+999999
- Small terminal + huge delta = invalid PTY size

**Fix**: Apply same clamping as height mode

---

### Bug: SIGWINCH Upper Bound Missing
**Status**: BROKEN
**Location**: main.go:726
**Severity**: MEDIUM

**Code**:
```go
if targetHeight < 1 {
    targetHeight = 1
}
// Missing: else if targetHeight > 9999 { targetHeight = 9999 }
```

**Issue**: Only lower bound checked
- DoD says 1-9999 range
- Large delta + terminal resize could exceed 9999
- uint16 allows up to 65535, so no natural limit

**Fix**: Add upper bound check

---

### Incomplete: test-toggle.sh Not Updated
**Status**: INCOMPLETE
**Location**: test-toggle.sh (unchanged)
**DoD**: Sprint 3 line 63

**Current**: Only tests original simple toggle (Ctrl+\×3)
**Missing**: Command mode UI, arrow keys, numeric entry

**Fix**: Update script with command mode examples or create new test-command-mode.sh

---

## Re-Evaluation Tracking

**Last Full Evaluation**: 2026-01-20 03:15:17
**Critical Bugs**: 3 (delta default, height validation, delta clamping)
**Incomplete Items**: 1 (test script)
**Cannot Verify**: 63+ interactive DoD criteria

**Files That Must Change**:
- main.go:373 (delta default)
- main.go:373-425 (add flag validation)
- main.go:672-673 (delta clamping)
- main.go:726 (SIGWINCH upper bound)
- test-toggle.sh (update with command mode)

**Re-test After Fixes**:
1. `-height N` should report N
2. `-height 99999` should error or clamp
3. Arrow keys in delta mode should clamp
4. User interactive testing session required
