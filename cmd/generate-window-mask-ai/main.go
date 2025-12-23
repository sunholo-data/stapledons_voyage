// Generate window masks using Gemini AI segmentation
// Production tool replacing threshold-based detection with AI vision
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
	"os"
	"path/filepath"
	"sort"
	"strings"

	"stapledons_voyage/engine/handlers"
)

// SegmentationResult from Gemini API
type SegmentationResult struct {
	Box2D   [4]int  `json:"box_2d"`  // [ymin, xmin, ymax, xmax] normalized 0-1000
	Polygon [][]int `json:"polygon"` // [[x,y], ...] normalized 0-1000
	Label   string  `json:"label"`
}

// DeckType defines prompts for different deck types
type DeckType string

const (
	DeckObservation DeckType = "observation"
	DeckBridge      DeckType = "bridge"
	DeckGeneric     DeckType = "generic"
)

var deckPrompts = map[DeckType]string{
	DeckObservation: `This is a spaceship observation deck with a large transparent dome or panoramic window.
Detect the SKY/SPACE areas - the large open regions between structural supports where stars would be visible.
Do NOT detect the structural framework, railings, or equipment - only the transparent viewing areas.
There may be one large dome or multiple window sections.`,

	DeckBridge: `This is a spaceship bridge/command center with viewport windows.
Detect the WINDOW areas - viewscreens and viewports showing space/stars.
Bridge typically has 1-3 main viewport windows at the front.
Do NOT detect control panels, screens, or interior lighting - only windows to space.`,

	DeckGeneric: `This is a spaceship interior with windows or viewports.
Detect all transparent areas where space/stars should be visible.
Include observation windows, portholes, viewscreens, and any glass panels.
Do NOT detect interior lights, screens, or reflective surfaces.`,
}

func main() {
	// Flags
	outputPath := flag.String("o", "", "Output mask path (default: <input>_mask.png)")
	deckType := flag.String("deck", "generic", "Deck type: observation, bridge, generic")
	customPrompt := flag.String("prompt", "", "Custom prompt (overrides deck type)")
	overlay := flag.Bool("overlay", false, "Also output visualization overlay")
	jsonOut := flag.Bool("json", false, "Output detected regions as JSON")
	maxRegions := flag.Int("max-regions", 10, "Maximum number of regions to detect")
	verbose := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	inputPath := args[0]

	// Determine output path
	outPath := *outputPath
	if outPath == "" {
		ext := filepath.Ext(inputPath)
		outPath = strings.TrimSuffix(inputPath, ext) + "_mask.png"
	}

	// Load image
	if *verbose {
		log.Printf("Loading: %s", inputPath)
	}
	img, err := loadImage(inputPath)
	if err != nil {
		log.Fatalf("Failed to load image: %v", err)
	}
	bounds := img.Bounds()
	imgW, imgH := bounds.Dx(), bounds.Dy()
	if *verbose {
		log.Printf("Image size: %dx%d", imgW, imgH)
	}

	// Create AI handler
	ctx := context.Background()
	aiHandler, err := handlers.NewAIHandlerFromEnv(ctx)
	if err != nil {
		log.Fatalf("Failed to create AI handler: %v", err)
	}

	// Build prompt
	prompt := *customPrompt
	if prompt == "" {
		dt := DeckType(*deckType)
		if p, ok := deckPrompts[dt]; ok {
			prompt = p
		} else {
			prompt = deckPrompts[DeckGeneric]
		}
	}

	// Request segmentation
	if *verbose {
		log.Println("Requesting AI segmentation...")
	}
	results, err := requestSegmentation(aiHandler, inputPath, prompt, *maxRegions)
	if err != nil {
		log.Fatalf("Segmentation failed: %v", err)
	}

	if len(results) == 0 {
		log.Println("Warning: No regions detected. Try a different prompt or check the image.")
		// Create empty mask
		mask := image.NewGray(bounds)
		if err := saveImage(mask, outPath); err != nil {
			log.Fatalf("Failed to save mask: %v", err)
		}
		fmt.Printf("Created empty mask: %s\n", outPath)
		return
	}

	log.Printf("Detected %d region(s)", len(results))
	for i, r := range results {
		if *verbose {
			log.Printf("  [%d] %s: box=[%d,%d,%d,%d], vertices=%d",
				i, r.Label, r.Box2D[0], r.Box2D[1], r.Box2D[2], r.Box2D[3], len(r.Polygon))
		}
	}

	// Output JSON if requested
	if *jsonOut {
		jsonData, _ := json.MarshalIndent(results, "", "  ")
		jsonPath := strings.TrimSuffix(outPath, ".png") + ".json"
		if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
			log.Printf("Warning: Failed to write JSON: %v", err)
		} else {
			fmt.Printf("Regions JSON: %s\n", jsonPath)
		}
	}

	// Generate mask
	mask := image.NewGray(bounds)
	for i, result := range results {
		if len(result.Polygon) >= 3 {
			fillPolygon(mask, result.Polygon, imgW, imgH)
			if *verbose {
				log.Printf("  [%d] Applied polygon (%d vertices)", i, len(result.Polygon))
			}
		} else {
			// Fall back to bounding box
			fillBoundingBox(mask, result.Box2D, imgW, imgH)
			if *verbose {
				log.Printf("  [%d] Applied bounding box (no polygon)", i)
			}
		}
	}

	// Calculate coverage
	coverage := calculateCoverage(mask)

	// Save mask
	if err := saveImage(mask, outPath); err != nil {
		log.Fatalf("Failed to save mask: %v", err)
	}
	fmt.Printf("Mask: %s (%.1f%% coverage)\n", outPath, coverage*100)

	// Save overlay if requested
	if *overlay {
		overlayImg := createOverlay(img, mask)
		overlayPath := strings.TrimSuffix(outPath, ".png") + "_overlay.png"
		if err := saveImage(overlayImg, overlayPath); err != nil {
			log.Printf("Warning: Failed to save overlay: %v", err)
		} else {
			fmt.Printf("Overlay: %s\n", overlayPath)
		}
	}
}

func printUsage() {
	fmt.Println("Generate window masks using AI segmentation")
	fmt.Println()
	fmt.Println("Usage: generate-window-mask-ai [flags] <image.png>")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -o <path>        Output mask path (default: <input>_mask.png)")
	fmt.Println("  -deck <type>     Deck type: observation, bridge, generic (default: generic)")
	fmt.Println("  -prompt <text>   Custom segmentation prompt")
	fmt.Println("  -overlay         Also output visualization overlay")
	fmt.Println("  -json            Output detected regions as JSON")
	fmt.Println("  -max-regions <n> Maximum regions to detect (default: 10)")
	fmt.Println("  -v               Verbose output")
	fmt.Println()
	fmt.Println("Deck types:")
	fmt.Println("  observation  Large dome/panoramic windows (1 main region)")
	fmt.Println("  bridge       Command center viewports (1-3 regions)")
	fmt.Println("  generic      Any spaceship interior windows")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  generate-window-mask-ai assets/decks/observation.png")
	fmt.Println("  generate-window-mask-ai -deck bridge -overlay assets/decks/bridge.png")
	fmt.Println("  generate-window-mask-ai -o out/mask.png -v image.png")
}

func requestSegmentation(ai handlers.AIHandler, imagePath, prompt string, maxRegions int) ([]SegmentationResult, error) {
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("reading image: %w", err)
	}
	base64Img := base64.StdEncoding.EncodeToString(imgData)

	reqPrompt := fmt.Sprintf(`%s

Output a JSON array of segmentation results. For each detected region:
- "box_2d": [ymin, xmin, ymax, xmax] bounding box, coordinates normalized to 0-1000
- "polygon": array of [x, y] vertex pairs tracing the region boundary, normalized to 0-1000
- "label": brief descriptive label

Use 30-100 polygon vertices for accurate boundaries.
Detect at most %d regions.
Return ONLY the JSON array, no other text.`, prompt, maxRegions)

	request := handlers.AIRequest{
		System: `You are an expert at image segmentation for game asset creation.
Analyze spaceship interior images and identify transparent window/viewport regions.
Return precise polygon boundaries that can be used as transparency masks.
Coordinates use 0-1000 normalized scale where (0,0) is top-left.
Always return valid JSON arrays.`,
		ResponseMIMEType: "application/json",
		MaxOutputTokens:  16384,
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

	var fullText string
	for _, block := range aiResponse.Content {
		if block.Text != "" {
			fullText += block.Text
		}
	}

	// Clean up response
	fullText = strings.TrimSpace(fullText)
	fullText = strings.TrimPrefix(fullText, "```json")
	fullText = strings.TrimPrefix(fullText, "```")
	fullText = strings.TrimSuffix(fullText, "```")
	fullText = strings.TrimSpace(fullText)

	// Find JSON array
	start := strings.Index(fullText, "[")
	end := strings.LastIndex(fullText, "]")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("no JSON array in response: %s", truncate(fullText, 200))
	}

	var results []SegmentationResult
	if err := json.Unmarshal([]byte(fullText[start:end+1]), &results); err != nil {
		return nil, fmt.Errorf("parsing results: %w", err)
	}

	return results, nil
}

func fillPolygon(mask *image.Gray, polygon [][]int, imgW, imgH int) {
	if len(polygon) < 3 {
		return
	}

	points := make([]image.Point, 0, len(polygon))
	minY, maxY := imgH, 0

	for _, pt := range polygon {
		if len(pt) < 2 {
			continue
		}
		x := pt[0] * imgW / 1000
		y := pt[1] * imgH / 1000
		points = append(points, image.Point{X: x, Y: y})
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}

	if len(points) < 3 {
		return
	}

	// Scanline fill
	for y := minY; y <= maxY; y++ {
		var intersections []int
		n := len(points)
		for i := 0; i < n; i++ {
			j := (i + 1) % n
			p1, p2 := points[i], points[j]

			if (p1.Y <= y && p2.Y > y) || (p2.Y <= y && p1.Y > y) {
				t := float64(y-p1.Y) / float64(p2.Y-p1.Y)
				x := int(float64(p1.X) + t*float64(p2.X-p1.X))
				intersections = append(intersections, x)
			}
		}

		sort.Ints(intersections)

		for i := 0; i+1 < len(intersections); i += 2 {
			for x := intersections[i]; x <= intersections[i+1]; x++ {
				if x >= 0 && x < imgW && y >= 0 && y < imgH {
					mask.SetGray(x, y, color.Gray{Y: 255})
				}
			}
		}
	}
}

func fillBoundingBox(mask *image.Gray, box [4]int, imgW, imgH int) {
	ymin := box[0] * imgH / 1000
	xmin := box[1] * imgW / 1000
	ymax := box[2] * imgH / 1000
	xmax := box[3] * imgW / 1000

	for y := ymin; y < ymax; y++ {
		for x := xmin; x < xmax; x++ {
			if x >= 0 && x < imgW && y >= 0 && y < imgH {
				mask.SetGray(x, y, color.Gray{Y: 255})
			}
		}
	}
}

func createOverlay(original image.Image, mask *image.Gray) image.Image {
	bounds := original.Bounds()
	out := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			origR, origG, origB, _ := original.At(x, y).RGBA()
			maskVal := mask.GrayAt(x, y).Y

			if maskVal > 127 {
				// Cyan overlay for masked areas
				out.Set(x, y, color.RGBA{
					R: uint8(origR >> 9),
					G: uint8((origG>>8 + 200) / 2),
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

func calculateCoverage(mask *image.Gray) float64 {
	bounds := mask.Bounds()
	total := bounds.Dx() * bounds.Dy()
	masked := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if mask.GrayAt(x, y).Y > 127 {
				masked++
			}
		}
	}

	return float64(masked) / float64(total)
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
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
