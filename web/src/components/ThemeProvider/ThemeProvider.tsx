import { ReactNode, useEffect } from 'react';
import { useThemeStore } from '../../store/useThemeStore';

interface ThemeProviderProps {
  children: ReactNode;
}

function ThemeProvider({ children }: ThemeProviderProps) {
  const { theme, loadTheme } = useThemeStore();

  // Load theme from localStorage or system preference on mount
  useEffect(() => loadTheme(), [loadTheme]);

  // Apply theme to doc
  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark');
  }, [theme]);

  return (
    <>{children}</>
  );
}

export default ThemeProvider;
