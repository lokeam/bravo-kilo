import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';


interface LibrarySortState {
  sortCriteria: 'title' | 'publishDate' | 'author' | 'pageCount';
  sortOrder: 'asc' | 'desc';
  setSort: (criteria: 'title' | 'publishDate' | 'author' | 'pageCount', order: 'asc' | 'desc') => void;
}


const useStore = create<LibrarySortState>()(
  persist(
    (set, get) => ({
      sortCriteria: 'title',
      sortOrder: 'asc',
      setSort: (criteria, order) => set({ sortCriteria: criteria, sortOrder: order }),
    }),
    {
      name: 'library-sort', // unique name for local storage
      storage: createJSONStorage(() => localStorage), // using localStorage for persistence
      partialize: (state) => ({ sortCriteria: state.sortCriteria, sortOrder: state.sortOrder }), // persist only part of the state
    }
  )
);

export default useStore;
