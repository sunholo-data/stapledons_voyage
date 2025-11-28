package scenario

import (
	"time"

	"stapledons_voyage/sim_gen"
)

// Metrics captures performance data during scenario runs
type Metrics struct {
	StartTime    time.Time
	EndTime      time.Time
	TickCount    int
	DrawCmdCount int
}

// NewMetrics creates a new metrics tracker
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime: time.Now(),
	}
}

// RecordTick updates metrics after a simulation tick
func (m *Metrics) RecordTick(out sim_gen.FrameOutput) {
	m.TickCount++
	m.DrawCmdCount += len(out.Draw)
}

// Finalize marks the end of metrics collection
func (m *Metrics) Finalize() {
	m.EndTime = time.Now()
}

// Duration returns the total time elapsed
func (m *Metrics) Duration() time.Duration {
	return m.EndTime.Sub(m.StartTime)
}

// AvgDrawCmds returns average draw commands per tick
func (m *Metrics) AvgDrawCmds() float64 {
	if m.TickCount == 0 {
		return 0
	}
	return float64(m.DrawCmdCount) / float64(m.TickCount)
}
