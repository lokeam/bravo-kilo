import { useThemeStore } from '../../store/useThemeStore';

const SettingsItemThemeBtn = () => {
  const { theme, toggleTheme } = useThemeStore();

  return (
    <button
      className="h-11 justify-stretch"
      onClick={toggleTheme}
    >
      { theme === 'light' ? 'Switch to Dark Mode' : 'Switch to Light Mode' }
    </button>
  );
};

export default SettingsItemThemeBtn;
