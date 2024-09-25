import { useThemeStore } from '../../store/useThemeStore';

function SettingsItemThemeBtn() {
  const { theme, toggleTheme } = useThemeStore();

  return (
    <button
      className="h-11 justify-stretch bg-gray-200 text-black dark:bg-gray-800 dark:text-white border-2 dark:border-2 transition duration-500 ease-in-out hover:border-vivid-blue dark:hover:border-vivid-blue"
      onClick={toggleTheme}
    >
      {theme === 'light' ? 'Switch to Dark Mode' : 'Switch to Light Mode'}
    </button>
  );
}

export default SettingsItemThemeBtn;
