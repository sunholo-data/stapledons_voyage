# AI-Powered Window Mask Generation

**Status:** IMPLEMENTED (2025-12-23)

## Summary

Replace the current threshold-based window mask generation with Gemini 2.5's native segmentation capabilities. This provides pixel-accurate masks through AI vision instead of brightness heuristics.

## Implementation

### Components Created

| Component | Location | Description |
|-----------|----------|-------------|
| Production tool | `cmd/generate-window-mask-ai/main.go` | AI segmentation with polygon rasterization |
| Demo/POC | `cmd/demo-gemini-segmentation/main.go` | Testing tool with multiple modes |
| Skill script | `.claude/skills/asset-manager/scripts/generate_mask_ai.sh` | Wrapper for easy invocation |
| Pipeline script | `.claude/skills/asset-manager/scripts/generate_transparent_scene.sh` | Full generation + masking pipeline |

### Components Removed (Deprecated)

| Component | Reason |
|-----------|--------|
| `cmd/generate-window-mask/` | Threshold-based, unreliable |
| `cmd/validate-window-mask/` | Replaced by AI validation |

### Usage

```bash
# Single mask generation
bin/generate-window-mask-ai -deck observation assets/decks/observation.png

# Full pipeline
.claude/skills/asset-manager/scripts/generate_transparent_scene.sh observation "dome deck"

# Verify all masks
.claude/skills/asset-manager/scripts/verify_deck_masks.sh
```

## Problem

Current workflow has limitations:

| Issue | Current Approach | Problem |
|-------|------------------|---------|
| Window detection | Brightness threshold | Misses dim windows, catches bright non-window areas |
| Region filtering | Keep top N largest | Manual tuning required per image |
| Shape accuracy | Binary threshold | No understanding of actual window boundaries |
| Iteration | Human adjusts threshold | Slow, tedious, inconsistent |

**Example failures:**
- Threshold 150 catches control panel lights as "windows"
- Threshold 180 misses side viewports that are slightly dimmer
- Geodesic dome panels: should be one mask, but detected as many fragments

## Solution

Use Gemini 2.5's segmentation API which returns:
- **Bounding boxes**: `[ymin, xmin, ymax, xmax]` normalized to 0-1000
- **Segmentation masks**: Base64 PNG probability maps (0-255 per pixel)
- **Labels**: Descriptive text for each detected region

### Mask Format Details

```json
[
  {
    "box_2d": [120, 50, 400, 300],
    "mask": "data:image/png;base64,iVBOR...",
    "label": "main observation window"
  },
  {
    "box_2d": [450, 100, 600, 350],
    "mask": "data:image/png;base64,iVBOR...",
    "label": "side viewport"
  }
]
```

**Processing steps:**
1. Decode base64 PNG mask (grayscale probability map)
2. Resize mask to match bounding box dimensions
3. Binarize at threshold (default: 127)
4. Place in full-size image at bounding box coordinates
5. Composite all masks with OR operation

## Architecture

### New Components

```
cmd/demo-gemini-segmentation/     # POC (created)
  main.go                         # Test segmentation API

cmd/generate-window-mask-ai/      # Production tool (new)
  main.go                         # Main segmentation workflow

.claude/skills/asset-manager/scripts/
  generate_mask_ai.sh             # Skill wrapper (new)
  iterate_mask.sh                 # Update to use AI segmentation
```

### Workflow Comparison

**Current (threshold-based):**
```
image.png → brightness threshold → filter regions → mask.png
              ↓
        AI counts windows → set keep-top=N
```

**New (AI segmentation):**
```
image.png → Gemini 2.5 segmentation → composite masks → mask.png
              ↓
        Returns pixel-accurate masks directly
```

## Implementation Plan

### Phase 1: Core Segmentation Command

Create `cmd/generate-window-mask-ai/main.go`:

```go
// Flags
-prompt string   // Custom segmentation prompt
-threshold int   // Binarization threshold (0-255, default 127)
-out string      // Output mask path
-overlay         // Also output visualization overlay
-boxes           // Also output bounding boxes visualization
-json            // Output detected regions as JSON
```

**Prompt engineering:**
```
Detect all window, viewport, and transparent glass areas in this spaceship interior.
Include observation windows, portholes, viewscreens, and any glass panels where
space/stars should be visible.

Output segmentation masks in JSON format with box_2d, mask, and label keys.
```

### Phase 2: Asset Manager Integration

Update `asset-manager` skill scripts:

1. **`generate_mask_ai.sh`** - New script using AI segmentation
   ```bash
   generate_mask_ai.sh <image.png> [output_mask.png]
   # Uses demo-gemini-segmentation under the hood
   ```

2. **`iterate_mask.sh`** - Update to use AI segmentation
   - Remove threshold iteration logic
   - Use single AI call for segmentation
   - Add prompt refinement if quality is low

3. **`verify_deck_masks.sh`** - Update verification
   - Check mask quality using AI validation
   - Verify mask aligns with detected window regions

### Phase 3: Deprecate Threshold-Based Tool

1. Keep `cmd/generate-window-mask/` for backwards compatibility
2. Add deprecation warning pointing to AI version
3. Update all scripts to use AI version by default

## API Details

### Gemini 2.5 Segmentation Request

```go
request := handlers.AIRequest{
    System: "You are an expert at image segmentation...",
    ResponseMIMEType: "application/json",
    MaxOutputTokens: 8192,  // Masks can be large
    Messages: []ContentBlock{
        {Type: ContentTypeImage, ImageRef: "data:image/png;base64,..."},
        {Type: ContentTypeText, Text: segmentationPrompt},
    },
}
```

### Response Parsing

```go
type SegmentationResult struct {
    Box2D [4]int `json:"box_2d"` // [ymin, xmin, ymax, xmax] normalized 0-1000
    Mask  string `json:"mask"`   // base64 PNG probability map
    Label string `json:"label"`  // descriptive label
}
```

### Mask Compositing

```go
func applyMask(composite *image.Gray, mask image.Image, box [4]int, imgW, imgH int, threshold uint8) {
    // Convert normalized coords (0-1000) to pixels
    ymin := box[0] * imgH / 1000
    xmin := box[1] * imgW / 1000
    // ... resize mask to box size, binarize, composite
}
```

## Acceptance Criteria

1. **Accuracy**: AI masks cover >90% of actual window area with <10% false positives
2. **Automation**: No manual threshold tuning required
3. **Speed**: Single API call generates complete mask (<5 seconds)
4. **Compatibility**: Output format matches existing `_mask.png` convention
5. **Fallback**: Graceful degradation if API unavailable (use threshold method)

## Cost Considerations

- Gemini 2.5 Flash: ~$0.0001 per image for segmentation
- ~10 deck images per game version = ~$0.001 per full regeneration
- Negligible compared to image generation costs

## Testing

1. **POC validation**: Test `demo-gemini-segmentation` on existing deck images
2. **Comparison**: Side-by-side with threshold masks on same images
3. **Edge cases**:
   - Multiple windows (bridge: 3 viewports)
   - Single large window (observation dome)
   - Mixed (panels + windows)
   - Dark scenes (dim windows)

## Dependencies

- Gemini 2.5 Flash (or Pro) with segmentation support
- Vertex AI or Gemini API access
- Go image processing (stdlib `image`, `image/png`)

## Open Questions

1. **Polygon vs PNG masks**: Gemini can also return SVG polygons - faster but less precise. Test both?
2. **Confidence thresholds**: Default 127 for binarization - should we expose this?
3. **Multi-pass refinement**: If first pass misses windows, retry with adjusted prompt?

## References

- [Gemini 2.5 Segmentation Guide](https://ai.google.dev/gemini-api/docs/vision)
- [Roboflow Gemini Segmentation Tutorial](https://blog.roboflow.com/gemini-2-5-object-detection-segmentation/)
- [Simon Willison's Gemini Segmentation Demo](https://simonwillison.net/2025/Apr/18/gemini-image-segmentation/)
- Current threshold tool: `cmd/generate-window-mask/main.go`
- POC demo: `cmd/demo-gemini-segmentation/main.go`
