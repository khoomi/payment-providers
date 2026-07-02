package payproviders

import "fmt"

type Manager struct {
	providers       map[Name]Provider
	defaultProvider Name
}

func NewManager(defaultProvider Name) *Manager {
	return &Manager{
		providers:       make(map[Name]Provider),
		defaultProvider: defaultProvider,
	}
}

func (m *Manager) Register(p Provider) {
	if m == nil || p == nil {
		return
	}
	m.providers[p.Name()] = p
}

func (m *Manager) Get(name Name) (Provider, error) {
	if m == nil {
		return nil, fmt.Errorf("payment manager is not configured")
	}
	if p, ok := m.providers[name]; ok && p != nil {
		return p, nil
	}
	if p, ok := m.providers[m.defaultProvider]; ok && p != nil {
		return p, nil
	}
	return nil, fmt.Errorf("selected payment provider %v is not available", name)
}

func (m *Manager) Default() Provider {
	if m == nil {
		return nil
	}
	return m.providers[m.defaultProvider]
}

func (m *Manager) HasProvider(name Name) bool {
	if m == nil {
		return false
	}
	_, ok := m.providers[name]
	return ok
}