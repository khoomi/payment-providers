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
	// Empty name → default only. Never silently substitute another gateway when
	// the caller asked for a specific provider (e.g. flutterwave → paystack).
	if name == "" {
		name = m.defaultProvider
	}
	if p, ok := m.providers[name]; ok && p != nil {
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