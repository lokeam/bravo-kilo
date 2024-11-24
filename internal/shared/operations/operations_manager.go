package operations

type Manager struct {
	Cache       *CacheOperation
	Domain      *DomainOperation
	Processor   *ProcessorOperation
}

// Responsibilities:

func NewManager(cache *CacheOperation, domain *DomainOperation, processor *ProcessorOperation) *Manager {
	return &Manager{
		Cache:       cache,
		Domain:      domain,
		Processor:   processor,
	}
}

