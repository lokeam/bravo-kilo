function SettingsItemAnimationBtn() {
  return(
    <label className="justify-center items-center cursor-pointer grid w-full">
      <input
        type="checkbox"
        value=""
        className="sr-only peer"
      />
      <div className="relative w-14 h-7 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-majorelle rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:start-[4px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-6 after:w-6 after:transition-all dark:border-gray-600 peer-checked:bg-majorelle"></div>
    </label>
  )
}

export default SettingsItemAnimationBtn;
