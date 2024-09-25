import { useState } from 'react';
import { useLocation, useNavigate, Link } from 'react-router-dom';
import { useAuth } from "../AuthContext";
import { useFocusContext  } from '../FocusProvider/FocusProvider';
import { useThemeStore } from '../../store/useThemeStore';
import Avatar from '../Avatar/Avatar';
import Modal from '../Modal/Modal';
import AutoComplete from '../AutoComplete/AutoComplete';

import { IoSearchOutline } from 'react-icons/io5';
import { IoMdSettings } from "react-icons/io";
import { MdLogout } from "react-icons/md";
import { ImArrowLeft2 } from "react-icons/im";

function TopNavigation() {
  const [opened, setOpened] = useState(false);
  const { logout } = useAuth();
  const { searchFocusRef } = useFocusContext();
  const { theme } = useThemeStore();
  const location = useLocation();
  const navigate = useNavigate();
  const isBookDetailOrSettingsPage = location.pathname.includes('library/books') || location.pathname.includes('settings');
  const isSearchPage = location.pathname.includes('/search');
  const isBooksRoute = location.pathname.includes('/books');
  const isAddPage = location.pathname.includes('/add');
  const isDetailPage = isBooksRoute && !isAddPage && !isSearchPage;
  const isLightTheme = theme === 'light'

  const openModal = () => setOpened(true);
  const closeModal = () => setOpened(false);

  const handleSettingsClick = () => {
    navigate(`/settings`);
    closeModal();
  };
  console.log('--------------');
  console.log('testing isDetailPage: ', isDetailPage);

  //<nav className={`${!isSearchPage ? 'dark:opacity-75' : ''}
  return (
    <header className="antialiased relative w-full h-auto">
      <nav className={`${isLightTheme && isDetailPage ? 'bg-transparent' : 'bg-white'} fixed border-none flex items-center content-between left-0 right-0 top-0 z-50 pr-8 lg:pr-12 h-[67px] text-white w-full dark:bg-black`}>
        <Link
          className={`${isBookDetailOrSettingsPage ? 'block' : 'hidden'} h-13 w-13 pl-7 pr-3 mr-6 inline-block content-center cursor-pointer bg-transparent border-none`}
          to={"/library"}
        >
          <ImArrowLeft2
            className="text-charcoal dark:text-white"
            size={20}
          />
        </Link>
        <div className={`relative flex flex-row  justify-items-center items-center w-full`}>

          {/* ----- Avatar / Nav Start ----- */}
          <div className={`${isBookDetailOrSettingsPage ? 'hidden' : 'block'} navLeft`}>
              <button
                className="flex border-none bg-transparent antialiased translate-x-0 mid:translate-x-0"
                onClick={openModal}
              >
                <Avatar />
              </button>
          </div>

          <Modal
            opened={opened}
            onClose={closeModal}
            title=""
          >
            <button
              className="flex flex-row justify-items-start items-center bg-transparent mr-1"
              onClick={handleSettingsClick}
            >
              <IoMdSettings
                className="mr-8"
                size={22}
              />
              <span>Settings</span>
            </button>
            <button
              className="flex flex-row justify-items-start items-center bg-transparent mr-1"
              onClick={logout}
            >
              <MdLogout
                className="mr-8"
                size={25}
              />
              <span>Log out</span>
            </button>
          </Modal>

          {/* ----- Search / Nav Center ----- */}
          <div className={`${isSearchPage ? 'visible' : 'invisible'} w-full`}>
            <AutoComplete />
          </div>

          {/* ----- Mobile / Nav End ----- */}
          <div className={`${isBookDetailOrSettingsPage ? 'hidden' : 'visible'} navEnd lg:invisible`}>
            <div className="flex items-center">
              <Link
                className="p-3 mr-1 text-gray-500 rounded hover:text-gray-900 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-white dark:hover:bg-gray-700 focus:ring-4 focus:ring-gray-300 dark:focus:ring-gray-600"
                to={"/library/books/search"}
                ref={ searchFocusRef }
              >
                <IoSearchOutline size={20} className="text-black dark:text-az-white"/>
              </Link>
            </div>
          </div>
        </div>
      </nav>
    </header>
  );
}

export default TopNavigation;
