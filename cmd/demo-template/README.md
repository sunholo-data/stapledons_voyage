# Demo Template

This directory contains a template for creating new demo binaries.

## Usage

1. Copy the template:
   ```bash
   cp -r cmd/demo-template cmd/demo-YOURNAME
   mv cmd/demo-YOURNAME/TEMPLATE.go cmd/demo-YOURNAME/main.go
   ```

2. Edit `main.go`:
   - Replace all `YOURNAME` with your demo name
   - Replace all `DESCRIPTION` with what the demo tests/shows
   - Update `outputPath` default to match your demo name
   - Implement your demo logic

3. Build and run:
   ```bash
   go build -o bin/demo-YOURNAME ./cmd/demo-YOURNAME
   bin/demo-YOURNAME
   ```

## Required Features (DO NOT REMOVE)

All demos MUST include:

- `--screenshot N` flag: Takes screenshot after N frames
- `--output PATH` flag: Custom screenshot output path
- `takeScreenshot()` function: Saves PNG to out/screenshots/
- Frame counter in HUD

These are required for automated visual testing and CI.

## Screenshot Verification

**IMPORTANT**: Always verify demos with screenshots before marking complete:

```bash
# Take screenshot after 60 frames
bin/demo-YOURNAME --screenshot 60

# View the screenshot
open out/screenshots/demo-YOURNAME.png
```

## Standard Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--screenshot` | 0 | Take screenshot after N frames (0=disabled) |
| `--output` | out/screenshots/demo-NAME.png | Screenshot output path |

Add your custom flags as needed.
