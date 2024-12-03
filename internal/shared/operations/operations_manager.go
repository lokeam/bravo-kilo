package operations

import (
	"context"
	"fmt"
	"sync"

	"github.com/lokeam/bravo-kilo/internal/shared/core"
	"github.com/lokeam/bravo-kilo/internal/shared/types"
)

// Manager needs to know which handlers it can process (Books, Movies, Games, etc)
type Manager struct {
	domains       map[core.DomainType]core.DomainHandler // Map of domain types to their handlers
	domainMutex   sync.RWMutex
	Cache         CacheOperator          // Use the interface instead of concrete type
	Domain        DomainOperator         // Define interface for domain operations
	Factory       *OperationFactory
}

// Interface for cache operations
type CacheOperator interface {
	Get(ctx context.Context, userID int, params *types.PageQueryParams) (any, error)
	Set(ctx context.Context, userID int, params *types.PageQueryParams, data any) error
}

func NewManager(
	cache CacheOperator,
	factory *OperationFactory,
) *Manager {
	if cache == nil {
		panic("cache cannot be nil")
	}
	if factory == nil {
		panic("factory cannot be nil")
	}

	return &Manager{
		domains:     make(map[core.DomainType]core.DomainHandler),
		Cache:       cache,
		Factory:     factory,
	}
}

func (m *Manager) ExecuteOperation(
	ctx context.Context,
	domain core.DomainType,
	pageType core.PageType,
	userID int,
	params *types.PageQueryParams,
) (interface{}, error) {
	operation, err := m.Factory.CreateOperation(domain, pageType)
	if err != nil {
		return nil, fmt.Errorf("failed to create operation: %w", err)
	}

	// Use the created operation
	data, err := operation.GetData(ctx, userID, params)
	if err != nil {
		return nil, fmt.Errorf("operation execution failed: %w", err)
	}

	return data, nil
}

func (m *Manager) RegisterDomain(handler core.DomainHandler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	m.domainMutex.Lock()
	defer m.domainMutex.Unlock()

	domainType := handler.GetType()
	if _, exists := m.domains[domainType]; exists {
		return fmt.Errorf("domain already registered: %v", domainType)
	}

	m.domains[domainType] = handler

	return nil
}
