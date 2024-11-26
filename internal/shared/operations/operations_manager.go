package operations

import (
	"fmt"
	"sync"

	"github.com/lokeam/bravo-kilo/internal/shared/core"
)

type Manager struct {
	domains       map[core.DomainType]core.DomainHandler
	domainMutex   sync.RWMutex
	Cache         *CacheOperation
	Domain        *DomainOperation
	Processor     *ProcessorOperation
}

func NewManager(cache *CacheOperation, domain *DomainOperation, processor *ProcessorOperation) *Manager {
	if cache == nil {
		panic("cache cannot be nil")
	}
	if domain == nil {
		panic("domain cannot be nil")
	}
	if processor == nil {
		panic("processor cannot be nil")
	}

	return &Manager{
		domains:     make(map[core.DomainType]core.DomainHandler),
		Cache:       cache,
		Domain:      domain,
		Processor:   processor,
	}
}

func (m *Manager) RegisterDomain(handler core.DomainHandler) error {
	if handler == nil {
		return fmt.Errorf("domain handler cannot be nil")
	}

	m.domainMutex.Lock()
	defer m.domainMutex.Unlock()

	domainType := handler.GetType()
	if _, exists := m.domains[domainType]; exists {
		return fmt.Errorf("domain handler already registered for domain type: %s", domainType)
	}

	m.domains[domainType] = handler
	return nil
}

// Return registered domain handler
func (m *Manager) GetDomainHandler(domainType core.DomainType) (core.DomainHandler, error) {
	m.domainMutex.RLock()
	defer m.domainMutex.RUnlock()

	handler, exists := m.domains[domainType]
	if !exists {
		return nil, fmt.Errorf("no handler registered for domain type %v", domainType)
	}

	return handler, nil
}
