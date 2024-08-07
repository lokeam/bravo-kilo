import { useState, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from "../AuthContext";
import Avatar from '../Avatar/Avatar';
import Modal from '../Modal/Modal';
import AutoComplete from '../AutoComplete/AutoComplete';

import useSearchStore from '../../store/useSearchStore';

import { IoSearchOutline } from 'react-icons/io5';
import { IoMdSettings } from "react-icons/io";
import { MdLogout } from "react-icons/md";
import { ImArrowLeft2 } from "react-icons/im";

export default function TopNavigation() {
  const [opened, setOpened] = useState(false);
  const { logout } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const isSearchPage = location.pathname.includes('library/books/search');

  const { addSearchHistory } = useSearchStore();

  const handleGoBack = () => {
    navigate(-1);
  };

  const handleSearchSubmit = (query: string) => {

    // Google Book Search api logic here
    // Example: fetchGoogleBooks(query);

    console.log('Search submitted: ', query)
    // add searchHistory query here

  };

  const openModal = () => setOpened(true);
  const closeModal = () => setOpened(false);

  return (
    <>
      <header className="antialiased relative w-full h-auto">
        <nav className="fixed border-none flex items-center content-center left-0 right-0 top-0 z-50 bg-black lg:px-6 h-[67px] text-white w-full">
          <button
            className={`${isSearchPage ? 'block' : 'hidden'} ml-1 mr-10 cursor-pointer bg-transparent border-none`}
            onClick={handleGoBack}
          >
            <ImArrowLeft2 size={20}/>
          </button>
          <div className={`relative flex flex-row  justify-items-center items-center w-full`}>

            {/* ----- Logo / Nav Start ----- */}
            <div className={`${isSearchPage ? 'hidden' : 'block'} navLeft`}>
                <button
                  className="flex border-none bg-transparent antialiased translate-x-0 mid:translate-x-0"
                  onClick={openModal}
                >
                  <Avatar />
                </button>
            </div>

            <Modal opened={opened} onClose={closeModal} title="">
              <button className="flex flex-row justify-items-start items-center bg-transparent mr-1">
                <IoMdSettings className="mr-8" size={22} />
                <span>Settings</span>
              </button>
              <button className="flex flex-row justify-items-start items-center bg-transparent mr-1" onClick={logout}>
                <MdLogout className="mr-8" size={25}/>
                <span>Log out</span>
              </button>
            </Modal>

            {/* ----- Search / Nav Center ----- */}
            <div className={`${isSearchPage ? 'visible' : 'invisible'} w-full`}>
              <AutoComplete onSubmit={handleSearchSubmit}/>
            </div>
            {/* <form onSubmit={(e) => e.preventDefault()} className={`${isSearchPage ? 'visible' : 'invisible'} w-full`} action="#">
              <label htmlFor="topbar-search" className="sr-only">Search</label>
              <div className="flex flex-row relative">
                <div className="flex absolute inset-y-0 left-0 items-center pl-3 pointer-events-none dark:text-white">
                  <IoSearchOutline />
                </div>
                <input
                  className="bg-maastricht border border-gray-600 text-az-white font-bold sm:text-sm rounded focus:ring-primary-500 focus:border-primary-500 block w-full pl-9 p-2.5 placeholder:polo-blue placeholder:font-bold"
                  id="topbar-search"
                  name="search"
                  onChange={handleSearchChange}
                  placeholder="Search for a book or author"
                  type="text"
                  value={searchQuery}
                />
              </div>
            </form> */}
            {/* ----- Mobile / Nav End ----- */}
            <div className={`${isSearchPage ? 'hidden' : 'visible'} navEnd lg:invisible`}>
              <div className="flex items-center">
                <button type="button" data-dropdown-toggle="notification-dropdown" className="p-3 mr-1 text-gray-500 rounded hover:text-gray-900 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-white dark:hover:bg-gray-700 focus:ring-4 focus:ring-gray-300 dark:focus:ring-gray-600">
                  <span className="sr-only">View notifications</span>
                    <IoSearchOutline />
                  </button>
              </div>
            </div>

          </div>
        </nav>
      </header>
    </>
  );
}
