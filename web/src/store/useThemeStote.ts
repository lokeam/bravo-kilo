import { create } from 'zustand';

interface ThemeState {
  theme: 'light' | 'dark';
  toggleTheme: () => void;
  setTheme: (theme: 'light' | 'dark') => void;
  loadTheme: () => void;
}

export const useThemeStore = create<ThemeState>((set) => ({
  // Set default theme
  theme: 'dark',

  toggleTheme: () =>
    set((state) => {
      const newTheme = state.theme === 'light' ? 'dark' : 'light';

      localStorage.setItem('theme', newTheme);
      return { theme: newTheme };
    }),

  setTheme: (theme: 'light' | 'dark') => set({ theme }),

  loadTheme: () => {
    const savedTheme = localStorage.getItem('theme') as 'light' | 'dark';
    if (savedTheme) {
      set({ theme: savedTheme });
    } else {
      // Check system theme preference
      const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      set({ theme: systemPrefersDark ? 'dark' : 'light' });
    }
  },
}));
