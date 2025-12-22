// Generate transparent windows from AI-generated deck images
// Supports multiple detection modes:
//   - black: Dark pixels (RGB < threshold) become transparent
//   - bright: Bright pixels (RGB > threshold), keep only N largest regions
//   - checker: Checkerboard pattern becomes transparent
//   - magenta: Magenta (#FF00FF) pixels become transparent
//
// Outputs either a mask OR the original image with transparency applied.
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"sort"
	"strings"

	"stapledons_voyage/engine/handlers"
)

func main() {
	// Flags
	mode := flag.String("mode", "black", "Detection mode: black, bright, checker, magenta")
	threshold := flag.Int("threshold", 32, "RGB threshold (0-255). black: below=window, bright: above=window")
	outputMask := flag.Bool("mask", false, "Output mask instead of transparent image")
	invert := flag.Bool("invert", false, "Invert detection (make detected areas opaque)")
	keepTop := flag.Int("keep-top", 0, "Keep only N largest regions (0=all, use with -mode=bright)")
	autoDetect := flag.Bool("auto-detect", false, "Use AI vision to automatically detect window count (bright mode only)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Usage: generate-window-mask [flags] <input.png> <output.png>")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		fmt.Println("\nModes:")
		fmt.Println("  black    - Detect dark pixels (RGB < threshold)")
		fmt.Println("  checker  - Detect checkerboard transparency pattern")
		fmt.Println("  magenta  - Detect magenta (#FF00FF) pixels")
		fmt.Println("\nExamples:")
		fmt.Println("  generate-window-mask deck.png deck_transparent.png")
		fmt.Println("  generate-window-mask -mode=checker frame.png frame_transparent.png")
		fmt.Println("  generate-window-mask -mode=magenta -mask scene.png mask.png")
		os.Exit(1)
	}

	inputPath := args[0]
	outputPath := args[1]

	// Load input image
	f, err := os.Open(inputPath)
	if err != nil {
		log.Fatalf("Failed to open input: %v", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatalf("Failed to decode image: %v", err)
	}

	// Auto-detect window count if requested
	if *autoDetect && *keepTop == 0 {
		log.Println("Auto-detecting window count with AI vision...")
		ctx := context.Background()
		aiHandler, err := handlers.NewAIHandlerFromEnv(ctx)
		if err != nil {
			log.Printf("Warning: Failed to create AI handler (%v), using default keep-top=3", err)
			*keepTop = 3
		} else {
			count, err := detectWindowCount(inputPath, aiHandler)
			if err != nil {
				log.Printf("Warning: AI detection failed (%v), using default keep-top=3", err)
				*keepTop = 3
			} else {
				log.Printf("AI detected %d window region(s)", count)
				*keepTop = count
			}
		}
	}

	bounds := img.Bounds()
	output := image.NewNRGBA(bounds)

	var detector func(x, y int, img image.Image) bool

	switch *mode {
	case "black":
		detector = func(x, y int, img image.Image) bool {
			r, g, b, _ := img.At(x, y).RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			return r8 < uint8(*threshold) && g8 < uint8(*threshold) && b8 < uint8(*threshold)
		}
	case "bright":
		detector = func(x, y int, img image.Image) bool {
			r, g, b, _ := img.At(x, y).RGBA()
			brightness := (int(r>>8) + int(g>>8) + int(b>>8)) / 3
			return brightness > *threshold
		}
	case "checker":
		detector = makeCheckerDetector(img)
	case "magenta":
		detector = func(x, y int, img image.Image) bool {
			r, g, b, _ := img.At(x, y).RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			// Allow some tolerance for magenta
			return r8 > 240 && g8 < 15 && b8 > 240
		}
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}

	width, height := bounds.Dx(), bounds.Dy()

	// Step 1: Build initial binary mask from detector
	binary := make([][]bool, height)
	for y := 0; y < height; y++ {
		binary[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			isWindow := detector(bounds.Min.X+x, bounds.Min.Y+y, img)
			if *invert {
				isWindow = !isWindow
			}
			binary[y][x] = isWindow
		}
	}

	// Step 2: If keep-top-N requested, filter to largest regions
	if *keepTop > 0 {
		binary = filterTopNRegions(binary, width, height, *keepTop)
	}

	// Step 3: Generate output from filtered binary mask
	windowPixels := 0
	totalPixels := width * height

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			isWindow := binary[y][x]
			imgX, imgY := bounds.Min.X+x, bounds.Min.Y+y

			if *outputMask {
				// Mask mode: white for windows, transparent elsewhere
				if isWindow {
					output.Set(imgX, imgY, color.NRGBA{255, 255, 255, 255})
					windowPixels++
				} else {
					output.Set(imgX, imgY, color.NRGBA{0, 0, 0, 0})
				}
			} else {
				// Transparency mode: keep original colors, make windows transparent
				r, g, b, _ := img.At(imgX, imgY).RGBA()
				r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

				if isWindow {
					output.Set(imgX, imgY, color.NRGBA{0, 0, 0, 0}) // Fully transparent
					windowPixels++
				} else {
					output.Set(imgX, imgY, color.NRGBA{r8, g8, b8, 255}) // Fully opaque
				}
			}
		}
	}

	// Save output
	outFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Failed to create output: %v", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, output); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}

	windowPercent := float64(windowPixels) / float64(totalPixels) * 100
	outputType := "transparent image"
	if *outputMask {
		outputType = "mask"
	}
	log.Printf("Created %s: %s", outputType, outputPath)
	log.Printf("Window coverage: %.1f%% (%d pixels)", windowPercent, windowPixels)
	log.Printf("Mode: %s", *mode)
}

// makeCheckerDetector creates a detector for checkerboard transparency patterns.
// The AI renders "transparency" as alternating light/dark gray squares.
func makeCheckerDetector(img image.Image) func(x, y int, img image.Image) bool {
	// Sample the image to find the checkerboard pattern colors
	// Typical AI checkerboard: ~204,204,204 (light) and ~153,153,153 (dark)
	lightGray := color.NRGBA{204, 204, 204, 255}
	darkGray := color.NRGBA{153, 153, 153, 255}

	return func(x, y int, img image.Image) bool {
		r, g, b, _ := img.At(x, y).RGBA()
		r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

		// Check if pixel is grayscale (R ≈ G ≈ B)
		if !isGrayscale(r8, g8, b8, 10) {
			return false
		}

		// Check if it matches either checkerboard color (with tolerance)
		tolerance := uint8(20)
		matchesLight := colorDistance(r8, g8, b8, lightGray.R, lightGray.G, lightGray.B) < tolerance
		matchesDark := colorDistance(r8, g8, b8, darkGray.R, darkGray.G, darkGray.B) < tolerance

		if !matchesLight && !matchesDark {
			return false
		}

		// Verify checkerboard pattern: neighbors should be opposite color
		// Check if this forms a checkerboard with 8x8 or similar grid
		bounds := img.Bounds()
		checkSize := 8 // Common checkerboard square size

		// Determine expected color based on position
		gridX := (x - bounds.Min.X) / checkSize
		gridY := (y - bounds.Min.Y) / checkSize
		expectLight := (gridX+gridY)%2 == 0

		if expectLight {
			return matchesLight
		}
		return matchesDark
	}
}

func isGrayscale(r, g, b uint8, tolerance uint8) bool {
	maxDiff := uint8(0)
	if diff := absDiff(r, g); diff > maxDiff {
		maxDiff = diff
	}
	if diff := absDiff(g, b); diff > maxDiff {
		maxDiff = diff
	}
	if diff := absDiff(r, b); diff > maxDiff {
		maxDiff = diff
	}
	return maxDiff <= tolerance
}

func colorDistance(r1, g1, b1, r2, g2, b2 uint8) uint8 {
	dr := absDiff(r1, r2)
	dg := absDiff(g1, g2)
	db := absDiff(b1, b2)
	return uint8(math.Sqrt(float64(dr*dr+dg*dg+db*db)) / 1.732) // Normalize to 0-255
}

func absDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

// filterTopNRegions keeps only the N largest connected regions in a binary mask.
func filterTopNRegions(binary [][]bool, width, height, keepTop int) [][]bool {
	// Connected component labeling
	labels := make([][]int, height)
	for y := 0; y < height; y++ {
		labels[y] = make([]int, width)
	}

	type region struct {
		label int
		size  int
	}
	var regions []region
	currentLabel := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if binary[y][x] && labels[y][x] == 0 {
				currentLabel++
				size := floodFill(binary, labels, x, y, width, height, currentLabel)
				regions = append(regions, region{currentLabel, size})
			}
		}
	}

	// Sort by size descending
	sort.Slice(regions, func(i, j int) bool {
		return regions[i].size > regions[j].size
	})

	// Keep top N labels
	keepLabels := make(map[int]bool)
	kept := 0
	for _, r := range regions {
		if kept >= keepTop {
			break
		}
		keepLabels[r.label] = true
		kept++
		log.Printf("Keeping region %d: %d pixels", kept, r.size)
	}

	log.Printf("Found %d regions total, keeping top %d", len(regions), kept)

	// Build filtered result
	result := make([][]bool, height)
	for y := 0; y < height; y++ {
		result[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			result[y][x] = keepLabels[labels[y][x]]
		}
	}

	return result
}

// floodFill labels connected bright pixels and returns region size
func floodFill(binary [][]bool, labels [][]int, startX, startY, width, height, label int) int {
	stack := []struct{ x, y int }{{startX, startY}}
	size := 0

	for len(stack) > 0 {
		// Pop
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		x, y := p.x, p.y

		// Bounds check
		if x < 0 || x >= width || y < 0 || y >= height {
			continue
		}

		// Skip if not bright or already labeled
		if !binary[y][x] || labels[y][x] != 0 {
			continue
		}

		// Label this pixel
		labels[y][x] = label
		size++

		// Add 4-connected neighbors
		stack = append(stack, struct{ x, y int }{x + 1, y})
		stack = append(stack, struct{ x, y int }{x - 1, y})
		stack = append(stack, struct{ x, y int }{x, y + 1})
		stack = append(stack, struct{ x, y int }{x, y - 1})
	}

	return size
}

// WindowCount is the structured output from AI vision analysis
type WindowCount struct {
	Count int    `json:"count"`
	Note  string `json:"note,omitempty"`
}

// detectWindowCount uses AI vision to count distinct window regions in a spaceship interior image
func detectWindowCount(imagePath string, aiHandler handlers.AIHandler) (int, error) {
	// Read and encode image to base64
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return 0, fmt.Errorf("reading image: %w", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)

	// Build structured AI request with vision using JSON schema for guaranteed valid output
	request := handlers.AIRequest{
		System: `You are analyzing spaceship interior scenes to count window regions for masking.

Count distinct window regions showing bright white light. If many small panels form one large dome, count as ONE region.`,
		ResponseMIMEType:   "application/json",
		ResponseJSONSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"count": map[string]interface{}{
					"type":        "integer",
					"description": "Number of distinct window regions (1-10)",
					"minimum":     1,
					"maximum":     10,
				},
				"note": map[string]interface{}{
					"type":        "string",
					"description": "Brief explanation of the count",
				},
			},
			"required": []string{"count", "note"},
		},
		MaxOutputTokens: 256, // Small JSON response only
		Messages: []handlers.ContentBlock{
			{
				Type:     handlers.ContentTypeImage,
				ImageRef: "data:image/png;base64," + base64Image,
				MimeType: "image/png",
			},
			{
				Type: handlers.ContentTypeText,
				Text: `Count the distinct window regions in this spaceship interior that show bright white light.

Rules:
- One large panoramic window = 1
- Three separate viewport windows = 3
- A geodesic dome made of many triangular panels = 1 (it's one dome)
- Two curved observation windows on the sides = 2

Return JSON only: {"count": N, "note": "brief explanation"}`,
			},
		},
	}

	// Encode request as JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return 0, fmt.Errorf("encoding request: %w", err)
	}

	// Call AI handler
	responseJSON, err := aiHandler.Call(string(requestJSON))
	if err != nil {
		return 0, fmt.Errorf("AI call failed: %w", err)
	}

	// Parse structured response
	var aiResponse handlers.AIResponse
	if err := json.Unmarshal([]byte(responseJSON), &aiResponse); err != nil {
		return 0, fmt.Errorf("parsing AI response: %w", err)
	}

	// Extract text from all content blocks (AI might split response)
	if len(aiResponse.Content) == 0 {
		return 0, fmt.Errorf("no content in AI response")
	}

	// Concatenate all text blocks in case response is split
	var fullText string
	for _, block := range aiResponse.Content {
		fullText += block.Text
	}

	// Try to extract JSON from response (in case it's wrapped in markdown or other text)
	responseText := fullText
	if strings.Contains(fullText, "{") && strings.Contains(fullText, "}") {
		start := strings.Index(fullText, "{")
		end := strings.LastIndex(fullText, "}") + 1
		responseText = fullText[start:end]
	}

	// Parse the structured JSON output from AI
	var result WindowCount
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		// Show truncated version for error message
		displayText := responseText
		if len(displayText) > 200 {
			displayText = displayText[:200] + "... (truncated)"
		}
		return 0, fmt.Errorf("parsing window count JSON: %w (response: %s)", err, displayText)
	}

	// Validate count
	if result.Count < 1 || result.Count > 10 {
		return 0, fmt.Errorf("invalid count %d (must be 1-10)", result.Count)
	}

	return result.Count, nil
}
