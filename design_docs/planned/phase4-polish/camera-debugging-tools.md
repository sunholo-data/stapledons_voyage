# Camera Debugging Tools

**Status**: Planned
**Target**: v0.2.0
**Priority**: P3 - Low
**Estimated**: 2-4 hours
**Dependencies**: [camera-lookat-fix.md](camera-lookat-fix.md)

## Game Vision Alignment

**Feature type:** Engine/Infrastructure - N/A on all pillars (enabling tech)

## Problem Statement

Recurring camera issues during demo development have been difficult to debug. We need better tooling to diagnose 3D camera problems quickly.

## Goals

Provide visual and logging tools to understand camera state during development.

## Proposed Tools

### 1. Camera State HUD Overlay
- Show camera position (X, Y, Z)
- Show camera rotation (as Euler angles)
- Show camera forward vector
- Show frustum near/far planes
- Toggle with F1 key

### 2. Debug Visualization
- Draw camera frustum wireframe
- Draw camera forward direction line
- Draw axis gizmo at camera position
- Toggle with F2 key

### 3. Enhanced Logging
- Log camera matrix changes
- Log when objects enter/leave view frustum
- Add `--debug-camera` CLI flag

### 4. Screenshot Comparison Script
- Capture before/after screenshots
- Diff images to highlight changes
- Output to `out/debug/` directory

## Implementation Notes

- Add to `engine/tetra/debug.go`
- Integrate with existing scene wrapper
- Should be compile-time optional (build tag)

## Non-Goals

- This is not the LookAt fix - see [camera-lookat-fix.md](camera-lookat-fix.md)
- Not real-time debugging UI (future feature)

---

**Document created**: 2025-12-08
**Last updated**: 2025-12-08
