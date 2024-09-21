import { useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import useStore from '../../store/useStore';

function LibraryNav() {
  const { activeTab, setActiveTab } = useStore();

  const handleTabClick = useCallback(
    (tab: string) => {
      console.log(`Switching to tab: ${tab}`);
      setActiveTab(tab);
    },
    [setActiveTab]
  );

  return (
    <div className="bookshelf_body relative w-full z-10 pb-8 box-border">
      <AnimatePresence mode="wait">
        <div className="bookshelf_grid_body box-content overflow-visible w-full">
          <ul className="bookshelf_grid_library text-left box-border grid grid-flow-col auto-cols-auto items-stretch gap-x-2.5 overflow-x-auto overflow-y-auto overscroll-x-none scroll-smooth snap-start snap-x snap-mandatory list-none m-0 pl-2 pb-5">
            {['All', 'Audiobooks', 'eBooks', 'Printed Books', 'Authors', 'Genres'].map((tab) => (
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
                    className="absolute -bottom-2 -left-1 rounded right-0 h-2 bg-majorelle"
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
