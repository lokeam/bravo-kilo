import { create } from 'zustand';
import { persist } from 'zustand/middleware';

// Match backed domain types
export type DomainType = 'books' | 'games' | 'movies' | 'music';

interface DomainState {
  // Current active todmain for API requests
  currentDomain: DomainType;

  // List of domains user has enabled
  activeDomains: DomainType[];

  // Actions
  setCurrentDomain: (domain: DomainType) => void;
  toggleDomain: (domain: DomainType) => void;
  isDomainActive: (domain: DomainType) => boolean;
}

export const useDomainStore = create<DomainState>()(
  persist(
    (set, get) => ({

      // Default domain is books
      currentDomain: 'books',
      // Initially only books is active
      activeDomains: ['books', 'games', 'movies', 'music'],

      // Set current domain used in DomainToggle component
      setCurrentDomain: (domain) =>
        set((state) => {
          // Only allow setting current domain if it's active
          if (!state.activeDomains.includes(domain)) {
            return state;
          }
          return { currentDomain: domain };
        }),

      // Toggle domain used in Settings page
      toggleDomain: (domain) => {
        set((state) => {
          // Don't allow deactivating the current domain
          if (domain === state.currentDomain) {
            return state;
          }

          const isActive = state.activeDomains.includes(domain);
          const newActiveDomains = isActive
            ? state.activeDomains.filter(d => d !== domain)
            : [...state.activeDomains, domain];

          return {
            activeDomains: newActiveDomains,
          };
        });
      },

      isDomainActive: (domain) => {
        return get().activeDomains.includes(domain);
      },
    }),
    {
      name: 'domain-storage',
      version: 1,
    }
  )
);