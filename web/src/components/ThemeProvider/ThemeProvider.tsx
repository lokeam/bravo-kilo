import { ReactNode, useEffect } from 'react';
import { useThemeStore } from '../../store/useThemeStore';

interface ThemeProviderProps {
  children: ReactNode;
}

function ThemeProvider({ children }: ThemeProviderProps) {
  const { theme, loadTheme } = useThemeStore();

  useEffect(() => {
    loadTheme();
  }, [loadTheme]);

  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark');
  }, [theme]);

  return <>{children}</>;
}

export default ThemeProvider;
