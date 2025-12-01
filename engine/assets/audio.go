// Package assets provides audio loading and playback functionality.
package assets

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

// SampleRate is the audio sample rate (44100 Hz is standard)
const SampleRate = 44100

// AudioManager handles sound effect and music playback.
type AudioManager struct {
	context   *audio.Context
	sounds    map[int][]byte          // Cached sound data by ID
	volumes   map[int]float64         // Volume per sound ID
	players   map[int]*audio.Player   // Active players for looping sounds
	bgmPlayer *audio.Player           // Background music player
	muted     bool
	volume    float64                 // Master volume (0.0 - 1.0)
}

// AudioManifest represents the sounds/manifest.json structure.
type AudioManifest struct {
	Sounds map[string]SoundEntry `json:"sounds"`
	BGM    map[string]BGMEntry   `json:"bgm"`
}

// SoundEntry defines a single sound effect in the manifest.
type SoundEntry struct {
	File   string  `json:"file"`
	Volume float64 `json:"volume"` // 0.0 - 1.0, defaults to 1.0
}

// BGMEntry defines background music in the manifest.
type BGMEntry struct {
	File   string  `json:"file"`
	Loop   bool    `json:"loop"`
	Volume float64 `json:"volume"`
}

// NewAudioManager creates a new audio manager.
func NewAudioManager() *AudioManager {
	return &AudioManager{
		context: audio.NewContext(SampleRate),
		sounds:  make(map[int][]byte),
		volumes: make(map[int]float64),
		players: make(map[int]*audio.Player),
		volume:  1.0,
	}
}

// LoadManifest loads sounds defined in the manifest.json file.
func (am *AudioManager) LoadManifest(soundPath string) error {
	manifestPath := filepath.Join(soundPath, "manifest.json")

	var manifest AudioManifest
	if err := loadJSON(manifestPath, &manifest); err != nil {
		return fmt.Errorf("loading audio manifest: %w", err)
	}

	// Load sound effects
	for idStr, entry := range manifest.Sounds {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			fmt.Printf("Warning: invalid sound ID %q, skipping\n", idStr)
			continue
		}

		soundPath := filepath.Join(soundPath, entry.File)
		data, err := am.loadSoundFile(soundPath)
		if err != nil {
			fmt.Printf("Warning: failed to load sound %d (%s): %v\n", id, entry.File, err)
			continue
		}

		am.sounds[id] = data
		am.volumes[id] = entry.Volume
		if am.volumes[id] == 0 {
			am.volumes[id] = 1.0 // Default volume
		}
	}

	return nil
}

// loadSoundFile loads and decodes a sound file (WAV or OGG).
func (am *AudioManager) loadSoundFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read file contents
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Decode based on extension
	ext := filepath.Ext(path)
	var decoded io.Reader

	switch ext {
	case ".wav":
		stream, err := wav.DecodeWithSampleRate(SampleRate, bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("decoding WAV: %w", err)
		}
		decoded = stream
	case ".ogg":
		stream, err := vorbis.DecodeWithSampleRate(SampleRate, bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("decoding OGG: %w", err)
		}
		decoded = stream
	default:
		return nil, fmt.Errorf("unsupported audio format: %s", ext)
	}

	// Read decoded audio into bytes
	decodedData, err := io.ReadAll(decoded)
	if err != nil {
		return nil, fmt.Errorf("reading decoded audio: %w", err)
	}

	return decodedData, nil
}

// PlaySound plays a sound effect by ID.
// Returns immediately - sound plays asynchronously.
func (am *AudioManager) PlaySound(id int) {
	if am.muted {
		return
	}

	data, ok := am.sounds[id]
	if !ok {
		// Sound not loaded - silently ignore (game should work without audio)
		return
	}

	// Create a new player for this sound
	player := audio.NewPlayerFromBytes(am.context, data)

	// Apply volume
	vol := am.volumes[id] * am.volume
	player.SetVolume(vol)

	// Play the sound
	player.Play()
}

// PlaySounds plays multiple sounds from FrameOutput.Sounds.
func (am *AudioManager) PlaySounds(soundIDs []int) {
	for _, id := range soundIDs {
		am.PlaySound(id)
	}
}

// SetVolume sets the master volume (0.0 - 1.0).
func (am *AudioManager) SetVolume(vol float64) {
	if vol < 0 {
		vol = 0
	}
	if vol > 1 {
		vol = 1
	}
	am.volume = vol
}

// GetVolume returns the current master volume.
func (am *AudioManager) GetVolume() float64 {
	return am.volume
}

// SetMuted mutes or unmutes all audio.
func (am *AudioManager) SetMuted(muted bool) {
	am.muted = muted
}

// IsMuted returns whether audio is muted.
func (am *AudioManager) IsMuted() bool {
	return am.muted
}

// HasSound returns true if a sound with the given ID is loaded.
func (am *AudioManager) HasSound(id int) bool {
	_, ok := am.sounds[id]
	return ok
}

// SoundCount returns the number of loaded sounds.
func (am *AudioManager) SoundCount() int {
	return len(am.sounds)
}
