// useStore.ts
import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

interface LibrarySortState {
  sortCriteria: 'title' | 'publishDate' | 'author' | 'pageCount';
  sortOrder: 'asc' | 'desc';
  setSort: (criteria: 'title' | 'publishDate' | 'author' | 'pageCount', order: 'asc' | 'desc') => void;
  activeTab: string;
  setActiveTab: (tab: string) => void;
}

const useStore = create<LibrarySortState>()(
  persist(
    (set, get) => ({
      sortCriteria: 'title',
      sortOrder: 'asc',
      setSort: (criteria, order) => {
        console.log(`Setting sort: ${criteria}, Order: ${order}`);
        set({ sortCriteria: criteria, sortOrder: order });
      },
      activeTab: 'All', // default active tab
      setActiveTab: (tab) => {
        console.log(`Changing active tab to: ${tab}`);
        set({ activeTab: tab });
      },
    }),
    {
      name: 'library-sort', // unique name for local storage
      storage: createJSONStorage(() => localStorage), // using localStorage for persistence
      partialize: (state) => ({
        sortCriteria: state.sortCriteria,
        sortOrder: state.sortOrder,
        activeTab: state.activeTab,
      }), // persist only part of the state
    }
  )
);

export default useStore;
