import { useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from "../AuthContext";
import Avatar from '../Avatar/Avatar';
import Modal from '../Modal/Modal';
import AutoComplete from '../AutoComplete/AutoComplete';


import { IoSearchOutline } from 'react-icons/io5';
import { IoMdSettings } from "react-icons/io";
import { MdLogout } from "react-icons/md";
import { ImArrowLeft2 } from "react-icons/im";
import { IoIosArrowDown } from "react-icons/io";


export default function TopNavigation() {
  const [opened, setOpened] = useState(false);
  const { logout } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const isSearchDetailPage = location.pathname.includes('library/books/');
  const isSearchPage = location.pathname.includes('library/books/search');

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
            className={`${isSearchDetailPage ? 'block' : 'hidden'} ml-1 mr-10 cursor-pointer bg-transparent border-none`}
            onClick={handleGoBack}
          >
            <ImArrowLeft2 size={20}/>
          </button>
          <div className={`relative flex flex-row  justify-items-center items-center w-full`}>

            {/* ----- Logo / Nav Start ----- */}
            <div className={`${isSearchDetailPage ? 'hidden' : 'block'} navLeft`}>
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
            {/* ----- Mobile / Nav End ----- */}
            <div className={`${isSearchDetailPage ? 'hidden' : 'visible'} navEnd lg:invisible`}>
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
