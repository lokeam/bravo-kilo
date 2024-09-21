import { useThemeStore } from '../../store/useThemeStore';

function SettingsItemThemeBtn() {
  const { theme, toggleTheme } = useThemeStore();

  return (
    <button
      className="h-11 justify-stretch bg-gray-200 text-black dark:bg-gray-800 dark:text-white"
      onClick={toggleTheme}
    >
      {theme === 'light' ? 'Switch to Dark Mode' : 'Switch to Light Mode'}
    </button>
  );
}

export default SettingsItemThemeBtn;
