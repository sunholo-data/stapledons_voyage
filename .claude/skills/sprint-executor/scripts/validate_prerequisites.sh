#!/usr/bin/env bash
# Validate prerequisites before starting sprint execution
# Adapted for Stapledons Voyage (AILANG game project)

set -euo pipefail

echo "Validating sprint prerequisites..."
echo

FAILURES=0
WARNINGS=0

# 1. Check working directory
echo "1/5 Checking working directory..."
if [[ -z $(git status --short 2>/dev/null) ]]; then
    echo "  ✓ Working directory clean"
else
    echo "  ⚠ Working directory has uncommitted changes:"
    git status --short | head -10
    WARNINGS=$((WARNINGS + 1))
fi
echo

# 2. Check AILANG modules compile
echo "2/5 Checking AILANG modules..."
AILANG_FAILURES=0
for f in sim/*.ail; do
    if [[ -f "$f" ]]; then
        if ailang check "$f" > /tmp/ailang_check.log 2>&1; then
            echo "  ✓ $f"
        else
            echo "  ✗ $f FAILS"
            cat /tmp/ailang_check.log
            AILANG_FAILURES=$((AILANG_FAILURES + 1))
        fi
    fi
done
if [[ $AILANG_FAILURES -gt 0 ]]; then
    echo "  ✗ $AILANG_FAILURES AILANG module(s) fail type-check"
    FAILURES=$((FAILURES + 1))
else
    echo "  ✓ All AILANG modules compile"
fi
echo

# 3. Check for AILANG messages
echo "3/5 Checking AILANG messages..."
INBOX_OUTPUT=$(ailang messages list --unread 2>&1 || echo "No messages")
if echo "$INBOX_OUTPUT" | grep -q "0 message"; then
    echo "  ✓ No pending messages"
elif echo "$INBOX_OUTPUT" | grep -q "message"; then
    MSG_COUNT=$(echo "$INBOX_OUTPUT" | grep -oP '\d+ message' | grep -oP '\d+' || echo "some")
    echo "  ⚠ $MSG_COUNT message(s) pending - check before starting"
    WARNINGS=$((WARNINGS + 1))
else
    echo "  ✓ Messages checked"
fi
echo

# 4. Check game build (if Makefile exists)
echo "4/5 Checking game build..."
if [[ -f "Makefile" ]]; then
    if make game > /tmp/game_build.log 2>&1; then
        echo "  ✓ Game builds successfully"
    else
        echo "  ⚠ Game build failed (may need engine work)"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    echo "  - No Makefile found, skipping build check"
fi
echo

# 5. Check CLAUDE.md for known limitations
echo "5/5 Checking known limitations..."
if [[ -f "CLAUDE.md" ]]; then
    echo "  Review known limitations before starting:"
    grep -A 10 "Known Limitations" CLAUDE.md 2>/dev/null | head -10 || echo "  (no limitations section found)"
else
    echo "  - No CLAUDE.md found"
fi
echo

# Summary
echo "================================"
if [[ $FAILURES -eq 0 ]]; then
    if [[ $WARNINGS -gt 0 ]]; then
        echo "✓ Prerequisites validated with $WARNINGS warning(s)"
        echo "Review warnings but can proceed with sprint."
    else
        echo "✓ All prerequisites validated!"
        echo "Ready to start sprint execution."
    fi
    exit 0
else
    echo "✗ $FAILURES prerequisite(s) failed"
    echo "Fix AILANG compilation errors before starting sprint."
    exit 1
fi
