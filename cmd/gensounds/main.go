// gensounds generates placeholder WAV sound files for testing.
// Creates simple sine wave tones at different frequencies.
package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

const sampleRate = 44100

// Sound definitions: name -> (frequency Hz, duration ms)
var sounds = map[string]struct {
	freq     float64
	duration int
}{
	"click.wav":  {800, 50},   // Short high click
	"build.wav":  {440, 200},  // Medium A note
	"error.wav":  {220, 300},  // Low warning tone
	"select.wav": {660, 100},  // Quick selection sound
}

func main() {
	outDir := "assets/sounds"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	for name, params := range sounds {
		path := filepath.Join(outDir, name)
		if err := generateWav(path, params.freq, params.duration); err != nil {
			fmt.Printf("Error generating %s: %v\n", name, err)
			continue
		}
		fmt.Printf("Generated %s (%.0f Hz, %d ms)\n", name, params.freq, params.duration)
	}

	fmt.Println("Done!")
}

// generateWav creates a WAV file with a sine wave tone.
func generateWav(path string, freq float64, durationMs int) error {
	numSamples := sampleRate * durationMs / 1000
	samples := make([]int16, numSamples*2) // Stereo

	// Generate sine wave with fade in/out
	fadeLen := numSamples / 10 // 10% fade
	for i := 0; i < numSamples; i++ {
		// Calculate amplitude with envelope
		amp := 0.3 // Max amplitude (30% to avoid clipping)

		// Fade in
		if i < fadeLen {
			amp *= float64(i) / float64(fadeLen)
		}
		// Fade out
		if i > numSamples-fadeLen {
			amp *= float64(numSamples-i) / float64(fadeLen)
		}

		// Generate sample
		t := float64(i) / float64(sampleRate)
		sample := int16(amp * 32767 * math.Sin(2*math.Pi*freq*t))

		// Stereo: same sample for both channels
		samples[i*2] = sample   // Left
		samples[i*2+1] = sample // Right
	}

	// Write WAV file
	return writeWav(path, samples, sampleRate)
}

// writeWav writes samples to a WAV file.
func writeWav(path string, samples []int16, sampleRate int) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// WAV header
	numChannels := 2
	bitsPerSample := 16
	byteRate := sampleRate * numChannels * bitsPerSample / 8
	blockAlign := numChannels * bitsPerSample / 8
	dataSize := len(samples) * 2 // 2 bytes per sample

	// RIFF header
	file.WriteString("RIFF")
	binary.Write(file, binary.LittleEndian, uint32(36+dataSize))
	file.WriteString("WAVE")

	// fmt chunk
	file.WriteString("fmt ")
	binary.Write(file, binary.LittleEndian, uint32(16))           // Chunk size
	binary.Write(file, binary.LittleEndian, uint16(1))            // Audio format (PCM)
	binary.Write(file, binary.LittleEndian, uint16(numChannels))  // Num channels
	binary.Write(file, binary.LittleEndian, uint32(sampleRate))   // Sample rate
	binary.Write(file, binary.LittleEndian, uint32(byteRate))     // Byte rate
	binary.Write(file, binary.LittleEndian, uint16(blockAlign))   // Block align
	binary.Write(file, binary.LittleEndian, uint16(bitsPerSample)) // Bits per sample

	// data chunk
	file.WriteString("data")
	binary.Write(file, binary.LittleEndian, uint32(dataSize))

	// Write samples
	for _, sample := range samples {
		binary.Write(file, binary.LittleEndian, sample)
	}

	return nil
}
