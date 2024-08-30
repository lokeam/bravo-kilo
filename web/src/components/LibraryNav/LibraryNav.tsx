import { useCallback } from 'react';
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
    <div className="bookshelf_body relative w-full z-10 pb-8">
      <div className="bookshelf_grid_wrapper box-border ">
        <div className="bookshelf_grid_body box-content overflow-visible w-full">
          <ul className="bookshelf_grid_library text-left box-border grid grid-flow-col auto-cols-auto items-stretch gap-x-2.5 overflow-x-auto overflow-y-auto overscroll-x-none scroll-smooth snap-start snap-x snap-mandatory list-none m-0 pb-5">
            {['All', 'Audiobooks', 'eBooks', 'Printed Books', 'Authors', 'Genres'].map((tab) => (
              <li
                key={tab}
                className={`flex items-center text-nowrap cursor-pointer ${
                  activeTab === tab ? 'text-3xl font-bold text-white' : 'text-lg font-semibold text-cadet-gray'
                }`}
                onClick={() => handleTabClick(tab)}
              >
                <span>{tab}</span>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  );
}

export default LibraryNav;
