// Package save provides single-file save/load for game state.
// Per Pillar 1 (Choices Are Final), only one save file exists - no save slots.
package save

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"stapledons_voyage/sim_gen"
)

// CurrentSaveVersion is the current save format version.
// Increment this when World schema changes.
const CurrentSaveVersion = "0.2.0" // 0.2.0: Added currentSystem (StarSystem)

// DefaultSavePath is the single save file location.
// No save slots - just one file that gets overwritten.
const DefaultSavePath = "saves/game.json"

// SaveFile represents the serialized game state.
type SaveFile struct {
	Version   string         `json:"version"`
	Timestamp int64          `json:"timestamp"`
	PlayTime  float64        `json:"play_time"` // Total seconds played
	World     *sim_gen.World `json:"world"`
}

// Manager handles save/load operations.
type Manager struct {
	savePath         string
	playTime         float64 // Accumulated play time
	lastSave         time.Time
	autoSaveInterval time.Duration
}

// NewManager creates a save manager with the default path.
func NewManager() *Manager {
	return &Manager{
		savePath:         DefaultSavePath,
		autoSaveInterval: 5 * time.Minute, // Auto-save every 5 minutes
	}
}

// NewManagerWithPath creates a save manager with a custom path.
func NewManagerWithPath(path string) *Manager {
	return &Manager{
		savePath:         path,
		autoSaveInterval: 5 * time.Minute,
	}
}

// SaveGame saves the current world state to the single save file.
// This overwrites any existing save - no slots, no branching (Pillar 1).
func (m *Manager) SaveGame(world *sim_gen.World) error {
	// Ensure directory exists
	dir := filepath.Dir(m.savePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating save directory: %w", err)
	}

	save := SaveFile{
		Version:   CurrentSaveVersion,
		Timestamp: time.Now().Unix(),
		PlayTime:  m.playTime,
		World:     world,
	}

	data, err := json.MarshalIndent(save, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling save: %w", err)
	}

	// Write to temp file first, then rename (atomic)
	tmpPath := m.savePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("writing save: %w", err)
	}

	if err := os.Rename(tmpPath, m.savePath); err != nil {
		os.Remove(tmpPath) // Clean up temp file
		return fmt.Errorf("finalizing save: %w", err)
	}

	m.lastSave = time.Now()
	return nil
}

// LoadGame loads the world state from the save file.
// Returns nil if no save exists (new game).
// Automatically migrates old saves to the current version.
func (m *Manager) LoadGame() (*sim_gen.World, error) {
	data, err := os.ReadFile(m.savePath)
	if os.IsNotExist(err) {
		return nil, nil // No save exists - new game
	}
	if err != nil {
		return nil, fmt.Errorf("reading save: %w", err)
	}

	var save SaveFile
	if err := json.Unmarshal(data, &save); err != nil {
		// Save is corrupted - backup and return error
		m.backupCorruptedSave()
		return nil, &CorruptedSaveError{
			Path:    m.savePath,
			Details: err.Error(),
		}
	}

	// Check if migration is needed
	if save.Version != CurrentSaveVersion {
		log.Printf("Save version %s differs from current %s, migrating...", save.Version, CurrentSaveVersion)
		if err := m.migrateWorld(save.World, save.Version); err != nil {
			m.backupCorruptedSave()
			return nil, &MigrationError{
				FromVersion: save.Version,
				ToVersion:   CurrentSaveVersion,
				Details:     err.Error(),
			}
		}
		log.Printf("Migration successful: %s -> %s", save.Version, CurrentSaveVersion)
	}

	// Restore play time
	m.playTime = save.PlayTime
	m.lastSave = time.Now()

	return save.World, nil
}

// migrateWorld upgrades an old World to the current schema.
func (m *Manager) migrateWorld(world *sim_gen.World, fromVersion string) (err error) {
	// Recover from panics during migration (AILANG codegen issues)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("migration panic: %v", r)
		}
	}()

	if world == nil {
		return fmt.Errorf("world is nil")
	}

	// Migration: 0.1.0 -> 0.2.0: Add currentSystem
	// NOTE: This migration is complex because the old save structure
	// is incompatible with new AILANG types. For now, fail migration
	// and let user start fresh.
	if fromVersion == "0.1.0" || fromVersion == "" {
		if world.CurrentSystem == nil {
			return fmt.Errorf("save version %s is too old - incompatible data structure. Please start a new game", fromVersion)
		}
	}

	// Future migrations go here:
	// if fromVersion == "0.2.0" { ... migrate to 0.3.0 ... }

	return nil
}

// backupCorruptedSave moves a corrupted save to a backup location.
func (m *Manager) backupCorruptedSave() {
	backupPath := m.savePath + ".corrupted." + time.Now().Format("20060102-150405")
	if err := os.Rename(m.savePath, backupPath); err != nil {
		log.Printf("Warning: failed to backup corrupted save: %v", err)
		return
	}
	log.Printf("Corrupted save backed up to: %s", backupPath)
}

// CorruptedSaveError indicates the save file could not be parsed.
type CorruptedSaveError struct {
	Path    string
	Details string
}

func (e *CorruptedSaveError) Error() string {
	return fmt.Sprintf("save file corrupted (%s): %s\nA backup has been created. Starting new game.", e.Path, e.Details)
}

// MigrationError indicates the save could not be migrated to the current version.
type MigrationError struct {
	FromVersion string
	ToVersion   string
	Details     string
}

func (e *MigrationError) Error() string {
	return fmt.Sprintf("failed to migrate save from v%s to v%s: %s\nA backup has been created. Starting new game.", e.FromVersion, e.ToVersion, e.Details)
}

// HasSave returns true if a save file exists.
func (m *Manager) HasSave() bool {
	_, err := os.Stat(m.savePath)
	return err == nil
}

// DeleteSave removes the save file (for "New Game" option).
func (m *Manager) DeleteSave() error {
	if err := os.Remove(m.savePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("deleting save: %w", err)
	}
	m.playTime = 0
	return nil
}

// UpdatePlayTime adds delta time to the accumulated play time.
// Call this each frame.
func (m *Manager) UpdatePlayTime(dt float64) {
	m.playTime += dt
}

// PlayTime returns the total play time in seconds.
func (m *Manager) PlayTime() float64 {
	return m.playTime
}

// ShouldAutoSave returns true if enough time has passed for an auto-save.
func (m *Manager) ShouldAutoSave() bool {
	return time.Since(m.lastSave) >= m.autoSaveInterval
}

// SetAutoSaveInterval sets how often auto-save triggers.
func (m *Manager) SetAutoSaveInterval(d time.Duration) {
	m.autoSaveInterval = d
}

// SavePath returns the current save file path.
func (m *Manager) SavePath() string {
	return m.savePath
}
