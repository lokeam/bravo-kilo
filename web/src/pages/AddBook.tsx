import { useState } from "react";
import axios from 'axios';
import { IoSearchOutline } from 'react-icons/io5';

const AddBook = () => {
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

  return (
    <section className="bg-gray-50 dark:bg-gray-900 h-full">
      <div className="flex flex-col items-center justify-center px-6 py-8 mx-auto h-screen lg:py-0">
        <div className="w-full bg-white rounded-lg shadow dark:border md:mt-0 sm:max-w-md xl:p-0 dark:bg-gray-800 dark:border-gray-700">
          <div className="p-6 space-y-4 md:space-y-6 sm:p-8">
            <h1 className="text-xl font-bold leading-tight tracking-tight text-gray-900 md:text-2xl dark:text-white">
              Add to Your Library
            </h1>
            <form onSubmit={handleSearchSubmit} className="space-y-4 md:space-y-6" action="#">
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
                  placeholder="Search for a book or author"
                  type="text"
                  value={searchQuery}
                />
              </div>
                <div className="flex items-center justify-between">
                  <div className="h-0.5 w-full bg-gray-600"></div>
                  <div className="text-center px-5">or</div>
                  <div className="h-0.5 w-full bg-gray-600"></div>
                </div>
                <button type="submit" className="w-full text-white bg-primary-600 hover:bg-primary-700 focus:ring-4 focus:outline-none focus:ring-primary-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center dark:bg-primary-600 dark:hover:bg-primary-700 dark:focus:ring-primary-800">Import a List of Books (.csv)</button>
                <button type="submit" className="w-full text-white bg-primary-600 hover:bg-primary-700 focus:ring-4 focus:outline-none focus:ring-primary-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center dark:bg-primary-600 dark:hover:bg-primary-700 dark:focus:ring-primary-800">Manually Add a Book</button>
            </form>
          </div>
        </div>
      </div>
    </section>
  );
}

export default AddBook;
