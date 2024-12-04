import { useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import useStore from '../../store/useStore';
import { useThemeStore } from '../../store/useThemeStore';

function LibraryNav() {
  const { activeTab, setActiveTab } = useStore();
  const { theme  } = useThemeStore()
  const isDarkTheme = theme == 'dark';
  console.log('testing theme: ', theme);


  const handleTabClick = useCallback(
    (tab: string) => {
      console.log(`Switching to tab: ${tab}`);
      console.log('Tab clicked:', {
        previous: activeTab,
        new: tab
      });
      setActiveTab(tab);
    },
    [setActiveTab, activeTab]
  );

  return (
    <div className={`${ isDarkTheme ? 'bookshelf_body_dk' : 'libNav_bsb_lt' } relative w-full z-10 pb-8 box-border`}>
      <AnimatePresence mode="wait">
        <div className="overflow-visible w-full">
          <ul className="bookshelf_grid_library text-left box-border grid grid-flow-col auto-cols-auto items-stretch gap-x-2.5 overflow-x-auto overflow-y-auto overscroll-x-none scroll-smooth snap-start snap-x snap-mandatory list-none m-0 pl-2 pb-5">
            {['All', 'Audiobooks', 'eBooks', 'Printed Books', 'Authors', 'Genres', 'Tags'].map((tab) => (
              <motion.li
                key={tab}
                className={`relative flex items-center text-nowrap cursor-pointer pb-2 ${
                  activeTab === tab ? 'text-3xl font-bold text-black dark:text-white' : 'text-lg font-semibold text-cadet-gray'
                }`}
                onClick={() => handleTabClick(tab)}
                animate
              >
                {tab}
                { activeTab === tab ? (
                  <motion.div
                    className="absolute -bottom-2 -left-1 rounded right-0 h-2 bg-vivid-blue"
                    layoutId="underline"
                  />
                ) : null}
              </motion.li>
            ))}
          </ul>
        </div>
      </AnimatePresence>
    </div>
  );
}

export default LibraryNav;
