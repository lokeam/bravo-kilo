import { create } from 'zustand';
import { persist, createJSONStorage, StateStorage } from 'zustand/middleware';

export interface SearchResult {
  id?: number | undefined;
  title: string;
  description: string;
  language: string;
  pageCount: number;
  publishDate: string;
  authors: string[];
  imageLinks: string[];
  genres: string[];
  isbn10: string;
  isbn13: string;
  isInLibrary: boolean;
}

export interface SearchEntry {
  timestamp: number;
  results: SearchResult[];
}

interface SearchStoreState {
  searchHistory: { [query: string]: SearchEntry };
  results: { [id: number]: SearchResult };
  addSearchHistory: (query: string, results: SearchResult[]) => void;
  clearSearchHistory: () => void;
  getFilteredSearchHistory: () => { [query: string]: SearchResult[] };
}

const FIVE_MINUTES_MS_EVECTION_LIMIT = 30 * 60 * 1000;
const customStorage: StateStorage = {
  getItem: (key) => {
    const value = localStorage.getItem(key);
    return value ? JSON.parse(value) : null;
  },
  setItem: (key, value) => {
    localStorage.setItem(key, JSON.stringify(value));
  },
  removeItem: (key) => {
    localStorage.removeItem(key);
  },
};

// Clear storage on app close
window.addEventListener('beforeunload', () => {
  localStorage.removeItem('search-history');
});

const useSearchStore = create<SearchStoreState>()(
  persist(
    (set, get) => ({
      searchHistory: {},
      results: {},
      addSearchHistory: (query, results) => {
        set((state) => {
          const newState = {
            searchHistory: {
              ...state.searchHistory,
              [query]: {
                timestamp: Date.now(),
                results,
              },
            },
            results: {
              ...state.results,
              ...results.reduce((acc, result) => {
                if (result.id !== undefined) {
                  acc[result.id] = result;
                }
                return acc;
              }, {} as { [id: number]: SearchResult }),
            },
          };

          console.log('Updated searchHistory:', newState.searchHistory);
          console.log('Updated results:', newState.results);

          return newState;
        });
      },
      clearSearchHistory: () => {
        set({ searchHistory: {}, results: {} });
        console.log('Search history cleared');
      },
      getFilteredSearchHistory: () => {
        const currentTime = Date.now();
        const validEntries: { [query: string]: SearchResult[] } = {};

        for (const [query, entry] of Object.entries(get().searchHistory)) {
          if (currentTime - entry.timestamp <= FIVE_MINUTES_MS_EVECTION_LIMIT) {
            validEntries[query] = entry.results;
          }
        }

        console.log('Filtered search history:', validEntries);

        return validEntries;
      },
    }),
    {
      name: 'search-history',
      storage: createJSONStorage(() => customStorage),
    }
  )
);

export default useSearchStore;
