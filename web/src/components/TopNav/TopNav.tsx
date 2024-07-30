import { useState } from 'react';
import { useAuth } from "../AuthContext";
import axios from 'axios';
import Avatar from '../Avatar/Avatar';
import Modal from '../Modal/Modal';

import { IoSearchOutline } from 'react-icons/io5';
import { IoMdSettings } from "react-icons/io";
import { MdLogout } from "react-icons/md";

export default function TopNavigation() {
  const [opened, setOpened] = useState(false);
  const { logout } = useAuth();

  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState([]);

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(e.target.value);
  };

  const handleSearchSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!searchQuery) return;

    try {
      const response = await axios.get(`${import.meta.env.VITE_API_ENDPOINT}/api/v1/books/search`, {
        params: { query: searchQuery },
        withCredentials: true,
      });
      setSearchResults(response.data.items);
      console.log('Search results:', response.data.items); // Add this line for debugging
    } catch (error) {
      console.error('Error fetching search results:', error);
    }
  };

  console.log('Search results state:', searchResults);

  const openModal = () => setOpened(true);
  const closeModal = () => setOpened(false);

  return (
    <>
      <header className="antialiased">
        <nav className="fixed left-0 right-0 top-0 z-50 bg-black px-4 lg:px-6 h-[67px] py-2.5 text-white">
          <div className="flex flex-row justify-between items-center">

            {/* ----- Logo / Nav Start ----- */}
            <div className="navLeft">
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
            <div className="navCenter hidden md:flex">
              <form onSubmit={handleSearchSubmit} className="hidden lg:block lg:pl-10 navCenterSearch">
                <label htmlFor="topbar-search" className="sr-only">Search</label>
                <div className="flex flex-row relative mt-1">
                  <div className="flex absolute inset-y-0 left-0 items-center pl-3 pointer-events-none dark:text-white">
                    <IoSearchOutline />
                  </div>
                  <input
                    className="bg-maastricht border border-gray-600 text-az-white font-bold sm:text-sm rounded focus:ring-primary-500 focus:border-primary-500 block w-full pl-9 p-2.5 placeholder:polo-blue placeholder:font-bold"
                    id="topbar-search"
                    name="search"
                    onChange={handleSearchChange}
                    placeholder="Search your library"
                    type="text"
                    value={searchQuery}
                  />
                </div>
              </form>
            </div>

            {/* ----- Mobile / Nav End ----- */}
            <div className="navEnd lg:invisible ">
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
