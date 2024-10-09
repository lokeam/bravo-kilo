import SettingsItem from '../components/SettingsItem/SettingsItem';
import PageWithErrorBoundary from '../components/ErrorMessages/PageWithErrorBoundary';

function Settings() {
  const settingsData = [
    {
      label: "Appearance",
      copy: "Customize how Bravo Kilo looks",
      variant: "theme",
    },
    {
      label: "Data Export",
      copy: "",
      variant: "export",
    },
    {
      label: "Animation",
      copy: "Disable animations and transitions",
      variant: "animation",
    },
    {
      label: "Delete Account",
      copy: "Permanently delete this account and all library data",
      variant: "delete",
    }
  ];

  return (
    <PageWithErrorBoundary fallbackMessage="Error loading settings page">
      <section className="bg-white dark:bg-black text-black dark:text-white relative flex flex-col items-center px-5 antialiased mdTablet:pr-5 mdTablet:ml-24 h-screen">
        <div className="text-left max-w-screen-mdTablet py-24 md:pb-4 flex flex-col relative w-full">
          <h2 className="mb-4 text-3xl font-bold text-black dark:text-white">Settings</h2>
          <div className="grid gap-4 grid-cols-1 sm:gap-6 py-3">
            {settingsData.map((item, index) => (
              <SettingsItem
                key={`${item.label}-${index}`}
                isLastItem={index === settingsData.length - 1 ? true : false}
                settingsData={item}
                variant={item.variant}
              />
            ))}
          </div>
        </div>
      </section>
    </PageWithErrorBoundary>
  );
}

export default Settings;
