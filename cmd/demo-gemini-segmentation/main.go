// Demo: Gemini 2.5 Segmentation for Window Mask Detection
// Uses Gemini's native segmentation capabilities to detect and mask window areas
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"stapledons_voyage/engine/handlers"
)

// SegmentationResult represents a single segmented object from Gemini
type SegmentationResult struct {
	Box2D   [4]int    `json:"box_2d"`  // [ymin, xmin, ymax, xmax] normalized 0-1000
	Mask    string    `json:"mask"`    // base64-encoded PNG probability map
	Polygon [][]int   `json:"polygon"` // Alternative: polygon vertices [[x,y], [x,y], ...] normalized 0-1000
	Label   string    `json:"label"`   // descriptive label
}

func main() {
	// Flags
	outputDir := flag.String("out", "out", "Output directory for results")
	threshold := flag.Int("threshold", 127, "Mask binarization threshold (0-255)")
	prompt := flag.String("prompt", "", "Custom segmentation prompt (default: detect windows)")
	showBoxes := flag.Bool("boxes", false, "Also output bounding boxes only (no masks)")
	usePolygon := flag.Bool("polygon", false, "Use polygon output instead of PNG masks (smaller, faster)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: demo-gemini-segmentation [flags] <image.png>")
		fmt.Println("\nFlags:")
		fmt.Println("  -out <dir>       Output directory (default: out)")
		fmt.Println("  -threshold <n>   Mask binarization threshold 0-255 (default: 127)")
		fmt.Println("  -prompt <text>   Custom segmentation prompt")
		fmt.Println("  -polygon         Use polygon mode (smaller, faster, recommended)")
		fmt.Println("  -boxes           Bounding boxes only (no mask data)")
		fmt.Println("\nModes (mutually exclusive):")
		fmt.Println("  default    PNG mask mode - base64 encoded probability maps (large)")
		fmt.Println("  -polygon   Polygon mode - vertex coordinates, rasterized locally (recommended)")
		fmt.Println("  -boxes     Bounding box mode - rectangles only, no pixel-level mask")
		fmt.Println("\nExamples:")
		fmt.Println("  demo-gemini-segmentation -polygon assets/decks/observation.png")
		fmt.Println("  demo-gemini-segmentation -polygon -prompt 'Detect sky/space areas' image.png")
		fmt.Println("  demo-gemini-segmentation -boxes assets/decks/bridge.png")
		os.Exit(1)
	}

	imagePath := args[0]

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output dir: %v", err)
	}

	// Load image
	log.Printf("Loading image: %s", imagePath)
	img, err := loadImage(imagePath)
	if err != nil {
		log.Fatalf("Failed to load image: %v", err)
	}
	bounds := img.Bounds()
	imgWidth, imgHeight := bounds.Dx(), bounds.Dy()
	log.Printf("Image size: %dx%d", imgWidth, imgHeight)

	// Create AI handler
	ctx := context.Background()
	aiHandler, err := handlers.NewAIHandlerFromEnv(ctx)
	if err != nil {
		log.Fatalf("Failed to create AI handler: %v", err)
	}

	// Build segmentation prompt
	segPrompt := *prompt
	if segPrompt == "" {
		segPrompt = `Detect all window, viewport, and transparent glass areas in this spaceship interior image.
These are areas where space/stars should be visible through the hull.
Include observation windows, portholes, viewscreens, and any glass panels.`
	}

	// Request segmentation
	log.Println("Requesting segmentation from Gemini 2.5...")
	results, err := requestSegmentation(aiHandler, imagePath, segPrompt, *showBoxes, *usePolygon)
	if err != nil {
		log.Fatalf("Segmentation failed: %v", err)
	}

	log.Printf("Found %d segmented regions", len(results))
	for i, r := range results {
		log.Printf("  [%d] %s: box=[%d,%d,%d,%d]", i, r.Label, r.Box2D[0], r.Box2D[1], r.Box2D[2], r.Box2D[3])
	}

	// Generate composite mask
	log.Println("Generating composite mask...")
	mask := image.NewGray(bounds)

	for i, result := range results {
		// Try polygon first (smaller, more reliable)
		if len(result.Polygon) > 2 {
			fillPolygon(mask, result.Polygon, imgWidth, imgHeight)
			log.Printf("  [%d] Applied polygon mask (%d vertices)", i, len(result.Polygon))
			continue
		}

		// Try PNG mask
		if result.Mask != "" {
			maskImg, err := decodeMaskPNG(result.Mask)
			if err != nil {
				log.Printf("  [%d] Failed to decode mask: %v", i, err)
			} else {
				applyMask(mask, maskImg, result.Box2D, imgWidth, imgHeight, uint8(*threshold))
				log.Printf("  [%d] Applied PNG mask", i)
				continue
			}
		}

		// Fall back to filling bounding box
		log.Printf("  [%d] No mask/polygon data, using bounding box", i)
		fillBoundingBox(mask, result.Box2D, imgWidth, imgHeight, uint8(*threshold+64))
	}

	// Save outputs
	baseName := strings.TrimSuffix(filepath.Base(imagePath), filepath.Ext(imagePath))

	// Save mask
	maskPath := filepath.Join(*outputDir, baseName+"_segmask.png")
	if err := saveImage(mask, maskPath); err != nil {
		log.Fatalf("Failed to save mask: %v", err)
	}
	log.Printf("Saved mask: %s", maskPath)

	// Save overlay visualization
	overlay := createOverlay(img, mask)
	overlayPath := filepath.Join(*outputDir, baseName+"_segoverlay.png")
	if err := saveImage(overlay, overlayPath); err != nil {
		log.Fatalf("Failed to save overlay: %v", err)
	}
	log.Printf("Saved overlay: %s", overlayPath)

	// Save bounding boxes visualization if requested
	if *showBoxes {
		boxesImg := drawBoundingBoxes(img, results, imgWidth, imgHeight)
		boxesPath := filepath.Join(*outputDir, baseName+"_boxes.png")
		if err := saveImage(boxesImg, boxesPath); err != nil {
			log.Printf("Failed to save boxes: %v", err)
		} else {
			log.Printf("Saved boxes: %s", boxesPath)
		}
	}

	// Print summary
	fmt.Println("\n=== Segmentation Results ===")
	fmt.Printf("Input: %s\n", imagePath)
	fmt.Printf("Regions detected: %d\n", len(results))
	fmt.Printf("Mask output: %s\n", maskPath)
	fmt.Printf("Overlay output: %s\n", overlayPath)
	fmt.Println("\nTo use this mask with the game:")
	fmt.Printf("  cp %s assets/decks/%s_mask.png\n", maskPath, baseName)
}

// requestSegmentation sends the image to Gemini and requests segmentation
func requestSegmentation(ai handlers.AIHandler, imagePath, prompt string, boxesOnly, usePolygon bool) ([]SegmentationResult, error) {
	// Load and encode image
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("reading image: %w", err)
	}
	base64Img := base64.StdEncoding.EncodeToString(imgData)

	// Build the request prompt
	var reqPrompt string
	if boxesOnly {
		reqPrompt = fmt.Sprintf(`%s

Output a JSON list of bounding boxes where each entry contains:
- "box_2d": [ymin, xmin, ymax, xmax] normalized to 0-1000
- "label": descriptive label for the object

Return ONLY the JSON array, no other text.`, prompt)
	} else if usePolygon {
		reqPrompt = fmt.Sprintf(`%s

Output a JSON list of segmentation polygons where each entry contains:
- "box_2d": [ymin, xmin, ymax, xmax] normalized to 0-1000
- "polygon": array of [x, y] coordinate pairs defining the mask boundary, normalized to 0-1000
- "label": descriptive label for the object

Use enough polygon vertices to capture the shape accurately (typically 20-100 points for complex shapes).
Return ONLY the JSON array, no other text.`, prompt)
	} else {
		reqPrompt = fmt.Sprintf(`%s

Output a JSON list of segmentation masks where each entry contains:
- "box_2d": [ymin, xmin, ymax, xmax] normalized to 0-1000
- "mask": base64 encoded PNG of the segmentation mask (probability map 0-255)
- "label": descriptive label for the object

Return ONLY the JSON array, no other text.`, prompt)
	}

	request := handlers.AIRequest{
		System: `You are an expert at image segmentation and object detection.
Analyze images and return precise bounding boxes and segmentation masks.
Always return valid JSON arrays. Coordinates use 0-1000 normalized scale.
Keep masks small by using appropriate compression.`,
		ResponseMIMEType: "application/json",
		MaxOutputTokens:  32768, // Masks can be very large
		Messages: []handlers.ContentBlock{
			{
				Type:     handlers.ContentTypeImage,
				ImageRef: "data:image/png;base64," + base64Img,
				MimeType: "image/png",
			},
			{
				Type: handlers.ContentTypeText,
				Text: reqPrompt,
			},
		},
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	responseJSON, err := ai.Call(string(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("AI call: %w", err)
	}

	// Parse response
	var aiResponse handlers.AIResponse
	if err := json.Unmarshal([]byte(responseJSON), &aiResponse); err != nil {
		return nil, fmt.Errorf("parsing AI response: %w", err)
	}

	// Extract text content
	var fullText string
	for _, block := range aiResponse.Content {
		if block.Text != "" {
			fullText += block.Text
		}
	}

	// Parse JSON array from response
	fullText = strings.TrimSpace(fullText)
	// Remove markdown code fences if present
	fullText = strings.TrimPrefix(fullText, "```json")
	fullText = strings.TrimPrefix(fullText, "```")
	fullText = strings.TrimSuffix(fullText, "```")
	fullText = strings.TrimSpace(fullText)

	var results []SegmentationResult
	if err := json.Unmarshal([]byte(fullText), &results); err != nil {
		// Try to find JSON array in response
		start := strings.Index(fullText, "[")
		end := strings.LastIndex(fullText, "]")
		if start >= 0 && end > start {
			if err2 := json.Unmarshal([]byte(fullText[start:end+1]), &results); err2 != nil {
				log.Printf("Raw response: %s", fullText[:min(500, len(fullText))])
				return nil, fmt.Errorf("parsing results JSON: %w (also tried: %w)", err, err2)
			}
		} else {
			log.Printf("Raw response: %s", fullText[:min(500, len(fullText))])
			return nil, fmt.Errorf("parsing results JSON: %w", err)
		}
	}

	return results, nil
}

// decodeMaskPNG decodes a base64-encoded PNG mask
func decodeMaskPNG(b64Data string) (image.Image, error) {
	// Handle data URI prefix if present
	if strings.HasPrefix(b64Data, "data:") {
		parts := strings.SplitN(b64Data, ",", 2)
		if len(parts) == 2 {
			b64Data = parts[1]
		}
	}

	decoded, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	img, err := png.Decode(bytes.NewReader(decoded))
	if err != nil {
		return nil, fmt.Errorf("PNG decode: %w", err)
	}

	return img, nil
}

// applyMask applies a segmentation mask to the composite at the bounding box location
func applyMask(composite *image.Gray, mask image.Image, box [4]int, imgW, imgH int, threshold uint8) {
	// Convert normalized coordinates (0-1000) to pixel coordinates
	ymin := box[0] * imgH / 1000
	xmin := box[1] * imgW / 1000
	ymax := box[2] * imgH / 1000
	xmax := box[3] * imgW / 1000

	boxW := xmax - xmin
	boxH := ymax - ymin

	maskBounds := mask.Bounds()
	maskW := maskBounds.Dx()
	maskH := maskBounds.Dy()

	// Sample mask and apply to composite
	for y := 0; y < boxH; y++ {
		for x := 0; x < boxW; x++ {
			// Map to mask coordinates
			mx := x * maskW / boxW
			my := y * maskH / boxH

			// Get mask value (grayscale)
			r, g, b, _ := mask.At(maskBounds.Min.X+mx, maskBounds.Min.Y+my).RGBA()
			val := uint8((r + g + b) / 3 >> 8)

			// Binarize at threshold
			if val >= threshold {
				// Set pixel in composite (white = masked area)
				composite.SetGray(xmin+x, ymin+y, color.Gray{Y: 255})
			}
		}
	}
}

// fillBoundingBox fills a bounding box region in the mask
func fillBoundingBox(mask *image.Gray, box [4]int, imgW, imgH int, value uint8) {
	ymin := box[0] * imgH / 1000
	xmin := box[1] * imgW / 1000
	ymax := box[2] * imgH / 1000
	xmax := box[3] * imgW / 1000

	for y := ymin; y < ymax; y++ {
		for x := xmin; x < xmax; x++ {
			mask.SetGray(x, y, color.Gray{Y: value})
		}
	}
}

// drawBoundingBoxes draws bounding boxes on the image
func drawBoundingBoxes(img image.Image, results []SegmentationResult, imgW, imgH int) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	// Copy original
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			out.Set(x, y, img.At(x, y))
		}
	}

	// Draw boxes
	boxColor := color.RGBA{0, 255, 0, 255} // Green
	for _, r := range results {
		ymin := r.Box2D[0] * imgH / 1000
		xmin := r.Box2D[1] * imgW / 1000
		ymax := r.Box2D[2] * imgH / 1000
		xmax := r.Box2D[3] * imgW / 1000

		// Draw rectangle outline
		for x := xmin; x <= xmax; x++ {
			out.Set(x, ymin, boxColor)
			out.Set(x, ymax, boxColor)
		}
		for y := ymin; y <= ymax; y++ {
			out.Set(xmin, y, boxColor)
			out.Set(xmax, y, boxColor)
		}
	}

	return out
}

// createOverlay creates a visualization with mask as red overlay
func createOverlay(original image.Image, mask *image.Gray) image.Image {
	bounds := original.Bounds()
	out := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			origR, origG, origB, _ := original.At(x, y).RGBA()
			maskVal := mask.GrayAt(x, y).Y

			if maskVal > 127 {
				// Blend with cyan for masked areas
				out.Set(x, y, color.RGBA{
					R: uint8(origR >> 9),       // Dim original
					G: uint8((origG>>8 + 200) / 2), // Add cyan
					B: uint8((origB>>8 + 255) / 2),
					A: 255,
				})
			} else {
				out.Set(x, y, color.RGBA{
					R: uint8(origR >> 8),
					G: uint8(origG >> 8),
					B: uint8(origB >> 8),
					A: 255,
				})
			}
		}
	}

	return out
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func saveImage(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// fillPolygon rasterizes a polygon onto the mask image
// Polygon coordinates are normalized 0-1000, converted to pixel coords
func fillPolygon(mask *image.Gray, polygon [][]int, imgW, imgH int) {
	if len(polygon) < 3 {
		return
	}

	// Convert normalized coords to pixels
	points := make([]image.Point, len(polygon))
	minY, maxY := imgH, 0
	for i, pt := range polygon {
		if len(pt) < 2 {
			continue
		}
		x := pt[0] * imgW / 1000
		y := pt[1] * imgH / 1000
		points[i] = image.Point{X: x, Y: y}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}

	// Scanline fill algorithm
	for y := minY; y <= maxY; y++ {
		// Find intersections with polygon edges
		var intersections []int
		n := len(points)
		for i := 0; i < n; i++ {
			j := (i + 1) % n
			p1, p2 := points[i], points[j]

			// Check if edge crosses this scanline
			if (p1.Y <= y && p2.Y > y) || (p2.Y <= y && p1.Y > y) {
				// Calculate x intersection
				t := float64(y-p1.Y) / float64(p2.Y-p1.Y)
				x := int(float64(p1.X) + t*float64(p2.X-p1.X))
				intersections = append(intersections, x)
			}
		}

		// Sort intersections
		sort.Ints(intersections)

		// Fill between pairs of intersections
		for i := 0; i+1 < len(intersections); i += 2 {
			for x := intersections[i]; x <= intersections[i+1]; x++ {
				if x >= 0 && x < imgW && y >= 0 && y < imgH {
					mask.SetGray(x, y, color.Gray{Y: 255})
				}
			}
		}
	}
}
