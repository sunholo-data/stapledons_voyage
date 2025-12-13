# Sprint: Voyage API Reference CLI

**Design Doc:** [design_docs/planned/voyage-api-reference.md](../design_docs/planned/voyage-api-reference.md)

## Overview

| Field | Value |
|-------|-------|
| Sprint ID | voyage-api-reference |
| Duration | 1 day |
| Priority | P2 (Developer tooling) |
| Status | ✅ Completed |

## Goal

Add `voyage api` CLI command that uses Go AST parsing to provide always-up-to-date engine API documentation. Since it parses source files at runtime, it automatically stays in sync with the latest code.

## Architecture Decision

**Using Option A: Go AST Parsing** (from design doc)

This approach:
- ✅ Always up-to-date with source (parses at runtime)
- ✅ Extracts doc comments automatically
- ✅ No build step or sync issues
- ✅ Simple implementation using `go/ast` and `go/parser`

## Tasks

### Phase 1: Core Implementation

- [x] Create `cmd/voyage/cmd_api.go` with AST parsing logic
- [x] Implement `extractPackageAPI()` function using `go/parser.ParseDir`
- [x] Extract exported types (structs, interfaces)
- [x] Extract exported functions and constructors
- [x] Extract methods (receiver functions)
- [x] Extract doc comments for all items

### Phase 2: Commands

- [x] Implement `voyage api` - list all engine packages
- [x] Implement `voyage api <package>` - list types in package (e.g., `voyage api tetra`)
- [x] Implement `voyage api <package>.<Type>` - show type details (e.g., `voyage api tetra.Scene`)
- [x] Implement `voyage api --search <query>` - search across all packages
- [x] Add `--methods` flag to show method signatures

### Phase 3: Integration

- [x] Add "api" case to main.go switch statement
- [x] Update printUsage() with api command documentation
- [x] Test with all engine packages (tetra, lod, shader, assets, render)
- [x] Verify output formatting is terminal-friendly

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `cmd/voyage/cmd_api.go` | ✅ Created | Main implementation |
| `cmd/voyage/main.go` | ✅ Modified | Add api command routing |

## Packages to Support

| Package | Path | Key Types |
|---------|------|-----------|
| tetra | `engine/tetra` | Scene, Planet, StarLight, AmbientLight |
| lod | `engine/lod` | Manager, Object, SimpleCamera, renderers |
| shader | `engine/shader` | Manager, SRWarp, GRWarp |
| assets | `engine/assets` | SpriteAtlas, AudioManager, FontManager |
| render | `engine/render` | DrawFrame, CaptureInput |
| camera | `engine/camera` | Camera transforms |
| display | `engine/display` | Window configuration |

## Success Criteria

From design doc:
- [x] `voyage api` lists all engine packages
- [x] `voyage api tetra` shows types in package
- [x] `voyage api tetra.Scene` shows constructor and methods
- [x] `voyage api --search "camera"` finds relevant APIs
- [x] Output is formatted for easy reading in terminal
- [x] Build with `go build -o bin/voyage ./cmd/voyage`

## Key Benefit: Always Up-to-Date

Since this uses runtime AST parsing of the actual source files:
- No manual documentation to maintain
- No JSON/YAML database to regenerate
- Adding/changing methods in engine automatically appears in `voyage api`
- Doc comments in source are the single source of truth

## Notes

- This is pure Go tooling - no AILANG involvement
- Performance target: < 1 second for full package scan
- Follows existing voyage CLI patterns (see cmd_demo.go, cmd_manifest.go)
