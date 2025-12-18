// Package textures provides GPU texture memory management with budget limits.
// This prevents runaway texture allocation from crashing the system.
package textures

import (
	"fmt"
	"image"
	"log"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
)

// Budget manages texture memory allocation with configurable limits.
// Use the global DefaultBudget or create custom budgets for different pools.
type Budget struct {
	maxBytes      int64 // Maximum allowed texture memory
	allocatedBytes int64 // Currently allocated (atomic)

	mu            sync.RWMutex
	allocations   map[*ebiten.Image]int64 // Track individual allocations

	// Callbacks
	onWarning     func(allocated, max int64)
	onLimitHit    func(requested, allocated, max int64)

	// Thresholds
	warningPct    float64 // Warn when this % of budget used (default 80%)
	warningLogged bool
}

// DefaultBudget is the global texture budget used by NewImageSafe.
// Initialized with sensible defaults based on system memory.
var DefaultBudget = NewBudget(0) // 0 = auto-detect

// NewBudget creates a new texture budget manager.
// If maxBytes is 0, auto-detects based on available system memory.
func NewBudget(maxBytes int64) *Budget {
	if maxBytes <= 0 {
		maxBytes = detectSafeTextureLimit()
	}

	b := &Budget{
		maxBytes:    maxBytes,
		allocations: make(map[*ebiten.Image]int64),
		warningPct:  0.80,
	}

	log.Printf("[textures] Budget initialized: %s max", formatBytes(maxBytes))
	return b
}

// detectSafeTextureLimit returns a safe texture memory limit based on system memory.
// Uses conservative defaults to leave room for other allocations.
func detectSafeTextureLimit() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get total system memory (approximate from Go's view)
	// We use Sys which is total memory obtained from OS
	totalMem := int64(m.Sys)

	// Also check GOGC memory limit if set
	if limit := debug.SetMemoryLimit(-1); limit > 0 && limit < totalMem {
		totalMem = limit
	}

	// If we can't determine, use a conservative 512MB default
	if totalMem < 100*1024*1024 {
		totalMem = 8 * 1024 * 1024 * 1024 // Assume 8GB system
	}

	// Use at most 25% of detected memory for textures
	// This leaves room for Go heap, OS, other apps
	safeLimit := totalMem / 4

	// Clamp to reasonable bounds
	minLimit := int64(128 * 1024 * 1024)  // 128MB minimum
	maxLimit := int64(2 * 1024 * 1024 * 1024) // 2GB maximum

	if safeLimit < minLimit {
		safeLimit = minLimit
	}
	if safeLimit > maxLimit {
		safeLimit = maxLimit
	}

	log.Printf("[textures] Detected system memory ~%s, texture limit set to %s",
		formatBytes(totalMem), formatBytes(safeLimit))

	return safeLimit
}

// SetMax updates the maximum texture budget.
func (b *Budget) SetMax(maxBytes int64) {
	b.mu.Lock()
	b.maxBytes = maxBytes
	b.mu.Unlock()
	log.Printf("[textures] Budget updated: %s max", formatBytes(maxBytes))
}

// Max returns the maximum texture budget in bytes.
func (b *Budget) Max() int64 {
	return atomic.LoadInt64(&b.maxBytes)
}

// Allocated returns currently allocated texture memory in bytes.
func (b *Budget) Allocated() int64 {
	return atomic.LoadInt64(&b.allocatedBytes)
}

// Available returns remaining texture budget in bytes.
func (b *Budget) Available() int64 {
	return b.Max() - b.Allocated()
}

// UsagePercent returns the percentage of budget currently used.
func (b *Budget) UsagePercent() float64 {
	max := b.Max()
	if max <= 0 {
		return 0
	}
	return float64(b.Allocated()) / float64(max) * 100
}

// SetWarningCallback sets a function called when usage exceeds warning threshold.
func (b *Budget) SetWarningCallback(fn func(allocated, max int64)) {
	b.mu.Lock()
	b.onWarning = fn
	b.mu.Unlock()
}

// SetLimitCallback sets a function called when allocation is rejected.
func (b *Budget) SetLimitCallback(fn func(requested, allocated, max int64)) {
	b.mu.Lock()
	b.onLimitHit = fn
	b.mu.Unlock()
}

// calcImageBytes estimates memory for an image of given dimensions.
// Assumes RGBA (4 bytes per pixel).
func calcImageBytes(width, height int) int64 {
	return int64(width) * int64(height) * 4
}

// CanAllocate returns true if the budget has room for an image of the given size.
func (b *Budget) CanAllocate(width, height int) bool {
	needed := calcImageBytes(width, height)
	return b.Available() >= needed
}

// NewImage creates a new ebiten.Image if within budget, tracking the allocation.
// Returns nil and an error if the budget would be exceeded.
func (b *Budget) NewImage(width, height int) (*ebiten.Image, error) {
	needed := calcImageBytes(width, height)
	allocated := b.Allocated()
	max := b.Max()

	// Check if allocation would exceed budget
	if allocated+needed > max {
		b.mu.RLock()
		cb := b.onLimitHit
		b.mu.RUnlock()

		if cb != nil {
			cb(needed, allocated, max)
		}

		return nil, fmt.Errorf("texture budget exceeded: need %s, have %s of %s",
			formatBytes(needed), formatBytes(max-allocated), formatBytes(max))
	}

	// Check warning threshold
	if !b.warningLogged && float64(allocated+needed)/float64(max) >= b.warningPct {
		b.mu.Lock()
		if !b.warningLogged {
			b.warningLogged = true
			log.Printf("[textures] WARNING: Budget %.0f%% used (%s of %s)",
				float64(allocated+needed)/float64(max)*100,
				formatBytes(allocated+needed), formatBytes(max))
			if b.onWarning != nil {
				b.onWarning(allocated+needed, max)
			}
		}
		b.mu.Unlock()
	}

	// Create the image
	img := ebiten.NewImage(width, height)

	// Track the allocation
	atomic.AddInt64(&b.allocatedBytes, needed)

	b.mu.Lock()
	b.allocations[img] = needed
	b.mu.Unlock()

	return img, nil
}

// NewImageFromImage creates an ebiten.Image from a Go image, tracking allocation.
func (b *Budget) NewImageFromImage(src image.Image) (*ebiten.Image, error) {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	needed := calcImageBytes(width, height)
	allocated := b.Allocated()
	max := b.Max()

	if allocated+needed > max {
		return nil, fmt.Errorf("texture budget exceeded: need %s, have %s of %s",
			formatBytes(needed), formatBytes(max-allocated), formatBytes(max))
	}

	img := ebiten.NewImageFromImage(src)

	atomic.AddInt64(&b.allocatedBytes, needed)

	b.mu.Lock()
	b.allocations[img] = needed
	b.mu.Unlock()

	return img, nil
}

// Release marks an image as freed, returning its memory to the budget.
// The image should no longer be used after calling this.
// Note: This doesn't actually dispose the GPU texture - that happens on GC.
func (b *Budget) Release(img *ebiten.Image) {
	if img == nil {
		return
	}

	b.mu.Lock()
	if bytes, ok := b.allocations[img]; ok {
		delete(b.allocations, img)
		atomic.AddInt64(&b.allocatedBytes, -bytes)
	}
	b.mu.Unlock()
}

// Stats returns current budget statistics.
type Stats struct {
	MaxBytes       int64
	AllocatedBytes int64
	AvailableBytes int64
	UsagePercent   float64
	AllocationCount int
}

// Stats returns current budget statistics.
func (b *Budget) Stats() Stats {
	b.mu.RLock()
	count := len(b.allocations)
	b.mu.RUnlock()

	allocated := b.Allocated()
	max := b.Max()

	return Stats{
		MaxBytes:        max,
		AllocatedBytes:  allocated,
		AvailableBytes:  max - allocated,
		UsagePercent:    float64(allocated) / float64(max) * 100,
		AllocationCount: count,
	}
}

// formatBytes formats bytes in human-readable form.
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// --- Convenience functions using DefaultBudget ---

// NewImageSafe creates a new ebiten.Image using the default budget.
// Returns nil and error if budget would be exceeded.
func NewImageSafe(width, height int) (*ebiten.Image, error) {
	return DefaultBudget.NewImage(width, height)
}

// NewImageFromImageSafe creates an ebiten.Image from a Go image using the default budget.
func NewImageFromImageSafe(src image.Image) (*ebiten.Image, error) {
	return DefaultBudget.NewImageFromImage(src)
}

// ReleaseImage marks an image as freed in the default budget.
func ReleaseImage(img *ebiten.Image) {
	DefaultBudget.Release(img)
}

// GetStats returns statistics from the default budget.
func GetStats() Stats {
	return DefaultBudget.Stats()
}

// SetMaxBudget updates the default texture budget limit.
func SetMaxBudget(maxBytes int64) {
	DefaultBudget.SetMax(maxBytes)
}

// CanAllocateImage returns true if the default budget can allocate an image of the given size.
func CanAllocateImage(width, height int) bool {
	return DefaultBudget.CanAllocate(width, height)
}
