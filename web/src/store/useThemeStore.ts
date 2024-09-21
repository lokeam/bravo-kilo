import { create } from 'zustand';

interface ThemeState {
  theme: 'light' | 'dark';
  toggleTheme: () => void;
  setTheme: (theme: 'light' | 'dark') => void;
  loadTheme: () => void;
}

export const useThemeStore = create<ThemeState>((set) => ({
  theme: 'light', // Set default theme to light since your styles are now light by default

  toggleTheme: () =>
    set((state) => {
      const newTheme = state.theme === 'light' ? 'dark' : 'light';
      localStorage.setItem('theme', newTheme);
      return { theme: newTheme };
    }),

  setTheme: (theme: 'light' | 'dark') => set({ theme }),

  loadTheme: () => {
    const savedTheme = localStorage.getItem('theme') as 'light' | 'dark' | null;
    if (savedTheme) {
      set({ theme: savedTheme });
    } else {
      const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      set({ theme: systemPrefersDark ? 'dark' : 'light' });
    }
  },
}));
