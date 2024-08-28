import SettingsItemThemeBtn from './SettingItemThemeBtn';
import SettingsItemExportBtn from './SettingItemExportBtn';
import SettingsItemAnimationBtn from './SettingItemAnimationBtn';
import SettingsItemDeleteAcctBtn from './SettingItemDeleteAcctBtn';

interface SettingsItemProps {
  settingsData: {
    label: string;
    copy: string;
  },
  isLastItem: boolean;
  variant: string;
  key: string;
}

function SettingsItem({ settingsData, isLastItem, variant }: SettingsItemProps) {
  return(
    <div className={`grid w-full gap-6 lgMobile:grid-cols-2 mdTablet:col-span-1 py-5 border-t ${isLastItem && 'border-b'} border-gray-600`}>
      <div className="block mb-2 text-base font-medium text-gray-900 dark:text-white">
        <h3 className="mb-1 text-xl font-bold">{settingsData.label}</h3>
        <p className="text-nevada-gray">{settingsData.copy}</p>
      </div>
      <div className="grid w-full">
        { variant === 'theme' && <SettingsItemThemeBtn /> }
        { variant === 'export' && <SettingsItemExportBtn /> }
        { variant === 'animation' && <SettingsItemAnimationBtn /> }
        { variant === 'delete' && <SettingsItemDeleteAcctBtn /> }
      </div>
    </div>
  )
}

export default SettingsItem;
