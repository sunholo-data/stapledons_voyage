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

	// Statistics
	stats Stats
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
	cameraPos := camera.Position()
	fovScale := camera.FOVScale()
	screenW := camera.ScreenWidth()
	screenH := camera.ScreenHeight()

	// Reset tier lists
	m.tier3D = m.tier3D[:0]
	m.tierBillboard = m.tierBillboard[:0]
	m.tierCircle = m.tierCircle[:0]
	m.tierPoint = m.tierPoint[:0]
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

		// Calculate apparent radius (screen-space size)
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

		// Assign tier based on distance
		obj.CurrentTier = m.calcTier(obj.Distance)
	}

	// Sort objects by distance for 3D priority
	sort.Slice(m.objects, func(i, j int) bool {
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

		switch obj.CurrentTier {
		case TierFull3D:
			// Limit 3D objects to Max3DObjects
			if num3D < m.config.Max3DObjects {
				m.tier3D = append(m.tier3D, obj)
				m.stats.Full3DCount++
				num3D++
			} else {
				// Demote to billboard if 3D pool is full
				obj.CurrentTier = TierBillboard
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

// calcTier determines the LOD tier based on distance.
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
