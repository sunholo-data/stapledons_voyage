package lod

import (
	"sort"
)

// Manager handles LOD tier assignment and object management.
type Manager struct {
	config  Config
	objects []*Object

	// Per-frame sorted lists by tier
	tier3D        []*Object
	tierBillboard []*Object
	tierCircle    []*Object
	tierPoint     []*Object

	// Objects currently transitioning (rendered in both tiers with blend)
	transitioning []*Object

	// Statistics
	stats Stats

	// Delta time for transition updates
	lastDT float64
}

// NewManager creates a new LOD manager with the given configuration.
func NewManager(config Config) *Manager {
	return &Manager{
		config:        config,
		objects:       make([]*Object, 0, 100),
		tier3D:        make([]*Object, 0, config.Max3DObjects),
		tierBillboard: make([]*Object, 0, 50),
		tierCircle:    make([]*Object, 0, 500),
		tierPoint:     make([]*Object, 0, 1000),
		transitioning: make([]*Object, 0, 20),
	}
}

// Add registers an object with the LOD manager.
func (m *Manager) Add(obj *Object) {
	m.objects = append(m.objects, obj)
}

// Remove unregisters an object by ID.
func (m *Manager) Remove(id string) {
	for i, obj := range m.objects {
		if obj.ID == id {
			// Remove by swapping with last element
			m.objects[i] = m.objects[len(m.objects)-1]
			m.objects = m.objects[:len(m.objects)-1]
			return
		}
	}
}

// Clear removes all objects.
func (m *Manager) Clear() {
	m.objects = m.objects[:0]
}

// ObjectCount returns the total number of managed objects.
func (m *Manager) ObjectCount() int {
	return len(m.objects)
}

// Update recalculates distances and assigns LOD tiers for all objects.
// This should be called once per frame before rendering.
func (m *Manager) Update(camera Camera) {
	m.UpdateWithDT(camera, 1.0/60.0) // Default 60 FPS
}

// UpdateWithDT updates with explicit delta time for smooth transitions.
func (m *Manager) UpdateWithDT(camera Camera, dt float64) {
	m.lastDT = dt
	cameraPos := camera.Position()
	fovScale := camera.FOVScale()
	screenW := camera.ScreenWidth()
	screenH := camera.ScreenHeight()

	// Reset tier lists
	m.tier3D = m.tier3D[:0]
	m.tierBillboard = m.tierBillboard[:0]
	m.tierCircle = m.tierCircle[:0]
	m.tierPoint = m.tierPoint[:0]
	m.transitioning = m.transitioning[:0]
	m.stats.Reset()
	m.stats.TotalObjects = len(m.objects)

	// Calculate distance and project for each object
	for _, obj := range m.objects {
		obj.Distance = cameraPos.Distance(obj.Position)

		// Project to screen space
		screenX, screenY, visible := camera.WorldToScreen(obj.Position)
		obj.ScreenX = screenX
		obj.ScreenY = screenY
		obj.Visible = visible

		// Calculate apparent radius (screen-space size in pixels)
		if obj.Distance > 0 {
			obj.ApparentRadius = (obj.Radius / obj.Distance) * fovScale
		} else {
			obj.ApparentRadius = fovScale // Very close, use max size
		}

		// Frustum culling: skip objects outside screen bounds
		margin := obj.ApparentRadius + 50 // Extra margin for large objects
		if screenX < -margin || screenX > float64(screenW)+margin ||
			screenY < -margin || screenY > float64(screenH)+margin {
			obj.Visible = false
		}

		// Calculate target tier with hysteresis
		targetTier := m.calcTierWithHysteresis(obj)

		// Handle tier transitions
		if targetTier != obj.CurrentTier {
			// Start new transition
			if obj.TransitionProgress >= 1.0 {
				obj.PreviousTier = obj.CurrentTier
				obj.TargetTier = targetTier
				obj.TransitionProgress = 0.0
			}
		}

		// Update transition progress
		if obj.TransitionProgress < 1.0 {
			if m.config.TransitionTime > 0 {
				obj.TransitionProgress += dt / m.config.TransitionTime
			} else {
				obj.TransitionProgress = 1.0 // Instant transition
			}
			if obj.TransitionProgress >= 1.0 {
				obj.TransitionProgress = 1.0
				obj.CurrentTier = obj.TargetTier
			}
		}
	}

	// Sort objects by distance (closest first) and importance
	sort.Slice(m.objects, func(i, j int) bool {
		// Higher importance always wins
		if m.objects[i].Importance != m.objects[j].Importance {
			return m.objects[i].Importance > m.objects[j].Importance
		}
		return m.objects[i].Distance < m.objects[j].Distance
	})

	// Distribute objects into tier lists
	num3D := 0
	for _, obj := range m.objects {
		if !obj.Visible {
			m.stats.CulledCount++
			continue
		}

		m.stats.VisibleCount++

		// Track transitioning objects separately
		if obj.IsTransitioning() {
			m.transitioning = append(m.transitioning, obj)
		}

		// Use target tier for bucket assignment during transition
		tier := obj.CurrentTier
		if obj.IsTransitioning() {
			tier = obj.TargetTier
		}

		switch tier {
		case TierFull3D:
			// Limit 3D objects to Max3DObjects
			if num3D < m.config.Max3DObjects || obj.Importance > 0 {
				m.tier3D = append(m.tier3D, obj)
				m.stats.Full3DCount++
				num3D++
			} else {
				// Demote to billboard if 3D pool is full
				obj.CurrentTier = TierBillboard
				obj.TargetTier = TierBillboard
				m.tierBillboard = append(m.tierBillboard, obj)
				m.stats.BillboardCount++
			}
		case TierBillboard:
			m.tierBillboard = append(m.tierBillboard, obj)
			m.stats.BillboardCount++
		case TierCircle:
			m.tierCircle = append(m.tierCircle, obj)
			m.stats.CircleCount++
		case TierPoint:
			m.tierPoint = append(m.tierPoint, obj)
			m.stats.PointCount++
		case TierCulled:
			m.stats.CulledCount++
		}
	}
}

// calcTierWithHysteresis determines the LOD tier with hysteresis to prevent flickering.
// Uses apparent size (pixels) when UseApparentSize is true, otherwise uses distance.
func (m *Manager) calcTierWithHysteresis(obj *Object) LODTier {
	apparentSize := obj.ApparentRadius
	currentTier := obj.CurrentTier
	hysteresis := m.config.Hysteresis

	if m.config.UseApparentSize {
		// Apparent size based (larger = more detail)
		// Upgrading (to higher detail): use normal threshold
		// Downgrading (to lower detail): use threshold * (1 - hysteresis)
		return m.calcTierByApparentSize(apparentSize, currentTier, hysteresis)
	}

	// Legacy distance based (smaller = more detail)
	return m.calcTierByDistance(obj.Distance, currentTier, hysteresis)
}

// calcTierByApparentSize determines tier based on screen pixel size.
func (m *Manager) calcTierByApparentSize(pixels float64, currentTier LODTier, hysteresis float64) LODTier {
	// Thresholds for upgrading to a tier
	full3D := m.config.Full3DPixels
	billboard := m.config.BillboardPixels
	circle := m.config.CirclePixels
	point := m.config.PointPixels

	// Apply hysteresis: harder to downgrade (need to be smaller)
	// Upgrading: pixels > threshold
	// Downgrading: pixels < threshold * (1 - hysteresis)
	downgradeMultiplier := 1.0 - hysteresis

	// Check from highest to lowest detail
	if pixels >= full3D || (currentTier == TierFull3D && pixels >= full3D*downgradeMultiplier) {
		return TierFull3D
	}
	if pixels >= billboard || (currentTier == TierBillboard && pixels >= billboard*downgradeMultiplier) {
		return TierBillboard
	}
	if pixels >= circle || (currentTier == TierCircle && pixels >= circle*downgradeMultiplier) {
		return TierCircle
	}
	if pixels >= point || (currentTier == TierPoint && pixels >= point*downgradeMultiplier) {
		return TierPoint
	}
	return TierCulled
}

// calcTierByDistance determines tier based on distance (legacy mode).
func (m *Manager) calcTierByDistance(distance float64, currentTier LODTier, hysteresis float64) LODTier {
	// For distance: smaller = closer = more detail
	// Upgrading: distance < threshold
	// Downgrading: distance > threshold * (1 + hysteresis)
	upgradeMultiplier := 1.0 + hysteresis

	if distance < m.config.Full3DDistance || (currentTier == TierFull3D && distance < m.config.Full3DDistance*upgradeMultiplier) {
		return TierFull3D
	}
	if distance < m.config.BillboardDistance || (currentTier == TierBillboard && distance < m.config.BillboardDistance*upgradeMultiplier) {
		return TierBillboard
	}
	if distance < m.config.CircleDistance || (currentTier == TierCircle && distance < m.config.CircleDistance*upgradeMultiplier) {
		return TierCircle
	}
	if distance < m.config.PointDistance || (currentTier == TierPoint && distance < m.config.PointDistance*upgradeMultiplier) {
		return TierPoint
	}
	return TierCulled
}

// calcTier determines the LOD tier based on distance (no hysteresis, for testing).
func (m *Manager) calcTier(distance float64) LODTier {
	switch {
	case distance < m.config.Full3DDistance:
		return TierFull3D
	case distance < m.config.BillboardDistance:
		return TierBillboard
	case distance < m.config.CircleDistance:
		return TierCircle
	case distance < m.config.PointDistance:
		return TierPoint
	default:
		return TierCulled
	}
}

// GetTier3D returns objects that should be rendered as full 3D meshes.
func (m *Manager) GetTier3D() []*Object {
	return m.tier3D
}

// GetTierBillboard returns objects that should be rendered as billboards.
func (m *Manager) GetTierBillboard() []*Object {
	return m.tierBillboard
}

// GetTierCircle returns objects that should be rendered as circles.
func (m *Manager) GetTierCircle() []*Object {
	return m.tierCircle
}

// GetTierPoint returns objects that should be rendered as points.
func (m *Manager) GetTierPoint() []*Object {
	return m.tierPoint
}

// GetTransitioning returns objects currently transitioning between tiers.
// These objects should be rendered in both their old and new tier with blending.
func (m *Manager) GetTransitioning() []*Object {
	return m.transitioning
}

// Stats returns the current LOD statistics.
func (m *Manager) Stats() Stats {
	return m.stats
}

// Config returns the current configuration.
func (m *Manager) Config() Config {
	return m.config
}

// SetConfig updates the LOD configuration.
func (m *Manager) SetConfig(config Config) {
	m.config = config
}

// GetObject returns an object by ID, or nil if not found.
func (m *Manager) GetObject(id string) *Object {
	for _, obj := range m.objects {
		if obj.ID == id {
			return obj
		}
	}
	return nil
}

// UpdatePosition updates an object's position by ID.
func (m *Manager) UpdatePosition(id string, pos Vector3) {
	if obj := m.GetObject(id); obj != nil {
		obj.Position = pos
	}
}
