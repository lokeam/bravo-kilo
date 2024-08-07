// src/store/useStore.ts

import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

interface SearchResult {
  id: number;
  text: string;
  subtitle?: string;
}

interface LibrarySortState {
  sortCriteria: 'title' | 'publishDate' | 'author' | 'pageCount';
  sortOrder: 'asc' | 'desc';
  setSort: (criteria: 'title' | 'publishDate' | 'author' | 'pageCount', order: 'asc' | 'desc') => void;
  activeTab: string;
  setActiveTab: (tab: string) => void;
  searchResults: SearchResult[];
  setSearchResults: (results: SearchResult[]) => void;
  searchHistory: Record<string, number[]>; // Cache structure for search history
  results: Record<number, SearchResult>; // Store search results by ID
  addSearchHistory: (query: string, results: SearchResult[]) => void; // Method to add to history
}

const useStore = create<LibrarySortState>()(
  persist(
    (set) => ({
      sortCriteria: 'title',
      sortOrder: 'asc',
      setSort: (criteria, order) => {
        console.log(`Setting sort: ${criteria}, Order: ${order}`);
        set({ sortCriteria: criteria, sortOrder: order });
      },
      activeTab: 'All',
      setActiveTab: (tab) => {
        console.log(`Changing active tab to: ${tab}`);
        set({ activeTab: tab });
      },
      searchResults: [],
      setSearchResults: (results) => {
        console.log('Setting search results');
        set({ searchResults: results });
      },
      searchHistory: {}, // Initialize search history
      results: {}, // Initialize results storage
      addSearchHistory: (query, newResults) => {
        set((state) => {
          const resultsCopy = { ...state.results };
          const resultIds: number[] = newResults.map((result) => {
            if (!resultsCopy[result.id]) {
              resultsCopy[result.id] = result;
            }
            return result.id;
          });

          return {
            results: resultsCopy,
            searchHistory: {
              ...state.searchHistory,
              [query]: resultIds,
            },
          };
        });
      },
    }),
    {
      name: 'library-sort',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        sortCriteria: state.sortCriteria,
        sortOrder: state.sortOrder,
        activeTab: state.activeTab,
        searchResults: state.searchResults,
        searchHistory: state.searchHistory, // Persist search history
        results: state.results, // Persist results
      }),
    }
  )
);

export default useStore;
