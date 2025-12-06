# Changelog

All notable changes to Stapledon's Voyage will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

**Source of truth:** GitHub tags and releases determine version numbers. See [Releases](https://github.com/sunholo-data/stapledons_voyage/releases).

## [Unreleased]

### Added
- Design doc validation script (`scripts/validate_design_docs.sh`)
- Feature summary generator (`scripts/generate_feature_summary.sh`)
- Post-release workflow improvements

### Changed
- Reorganized design documentation structure
- Updated design_docs/README.md with cleaner format

---

## [0.1.0] - 2025-12-06

Initial release with functional engine and relativistic visual effects.

### Added

#### Engine Foundation
- 2D rendering engine using Go/Ebiten
- Game loop with input handling
- Asset loading system (sprites, fonts, sounds)
- Display configuration and fullscreen support (F11)

#### Shader System
- Kage shader pipeline for post-processing
- Bloom, vignette, CRT scanlines, chromatic aberration effects
- Effect toggle system (F5-F9 keys)
- Design doc: [shader-system.md](design_docs/implemented/v0_1_0/shader-system.md)

#### Special Relativity Visual Effects
Physically accurate relativistic visuals when traveling at near-light speeds:
- **Aberration** - Stars compress into forward cone at high velocity
- **Doppler shift** - Forward stars blueshift, rear stars redshift
- **Relativistic beaming** - Forward brightens (D^3), rear dims to near-black
- Demo controls: F4 toggles, Shift+F4 cycles velocity (0.5c → 0.99c)
- Design doc: [sr-effects.md](design_docs/implemented/v0_1_0/sr-effects.md)

#### General Relativity Visual Effects
Gravitational lensing near massive objects (black holes, neutron stars):
- **Gravitational lensing** - Light bends around massive objects
- **Einstein ring** - Photon sphere creates bright halo at r = 1.5rs
- **Event horizon** - Central darkness where light cannot escape
- Design doc: [sprint-relativistic-effects.md](design_docs/implemented/v0_1_0/sprint-relativistic-effects.md)

#### Audio System
- OGG/WAV audio loading via manifest
- PlaySound API for game events
- Volume control foundation
- Design doc: [audio-system.md](design_docs/implemented/v0_1_0/audio-system.md)

#### Testing Infrastructure
- Headless screenshot capture for visual testing
- Visual scenario runner for automated tests
- Golden file comparison for regression detection
- Design docs: [screenshot-mode.md](design_docs/implemented/v0_1_0/screenshot-mode.md), [test-scenarios.md](design_docs/implemented/v0_1_0/test-scenarios.md)

#### AILANG Integration
- Mock sim_gen package with protocol types
- Foundation for AILANG-based simulation logic
- Effect handlers (Debug, Rand, Clock, AI stubs)

---

## Roadmap

See [design_docs/README.md](design_docs/README.md) for detailed feature planning.

### v0.2.0 - Gameplay Foundation (Planned)
- UI modes architecture
- Ship exploration mode
- Basic NPC rendering

### v0.3.0 - Core Systems (Planned)
- Galaxy map navigation
- Dialogue system
- Save/load system

### v0.4.0 - Journey System (Planned)
- Journey planning with time dilation
- Civilization simulation
- Trade mechanics

### v0.5.0+ - Endgame (Planned)
- Exploration modes
- Endgame legacy visualization
- Supporting UIs

---

[Unreleased]: https://github.com/sunholo-data/stapledons_voyage/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/sunholo-data/stapledons_voyage/releases/tag/v0.1.0
