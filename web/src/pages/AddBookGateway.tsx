import { useState } from 'react';
import { NavLink } from 'react-router-dom';
import { IoSearchOutline } from 'react-icons/io5';

import Snackbar from '../components/Snackbar/Snackbar';

const AddBookGateway = () => {
  const [snackbarOpen, setSnackbarOpen] = useState<boolean>(false);
  const [snackbarVariant, setSnackbarVariant] = useState<'added' | 'updated' | 'removed' | 'error'>('added');
  const [snackbarMsg, setSnackbarMsg] = useState<string>('');

  const handleShowSnackbar = (variant: 'added' | 'updated' | 'removed' | 'error', msg: string) => {
    setSnackbarOpen(true);
    setSnackbarVariant(variant);
    setSnackbarMsg(msg);
    setSnackbarOpen(true);
  };

  const handleCloseSnackbar = () => {
    setSnackbarOpen(false);
  }

  return (
    <section className="bg-gray-50 dark:bg-gray-900 px-5 mdTablet:ml-24 h-full pt-28">
      <div className="flex flex-col items-center py-8 mx-auto h-screen lg:py-0">
        <div className="w-full bg-white rounded-lg shadow dark:border md:mt-0 sm:max-w-md xl:p-0 dark:bg-gray-800 dark:border-gray-700">
          <div className="p-6 space-y-4 md:space-y-6 sm:p-8">
            <h1 className="text-xl font-bold leading-tight tracking-tight text-gray-900 md:text-2xl dark:text-white">
              Add to Your Library
            </h1>
            <div className="space-y-4 md:space-y-6">
              <label htmlFor="topbar-search" className="sr-only">Search</label>
              <div className="flex flex-row relative mt-1">
                <div className="flex absolute inset-y-0 left-0 items-center pl-3 pointer-events-none dark:text-white">
                  <IoSearchOutline />
                </div>
                <NavLink
                  className="bg-maastricht border border-gray-600 hover:text-az-white text-az-white font-bold sm:text-sm rounded  block w-full pl-9 p-2.5 placeholder:polo-blue placeholder:font-bold"
                  to={"/library/books/search"}
                  end
                >
                  Add book via Search
                </NavLink>
              </div>
                <div className="flex items-center justify-between">
                  <div className="h-0.5 w-full bg-gray-600"></div>
                  <div className="text-center px-5">or</div>
                  <div className="h-0.5 w-full bg-gray-600"></div>
                </div>
                <NavLink
                  className="block w-full text-white hover:text-white bg-primary-600 hover:bg-primary-700 focus:ring-4 focus:outline-none focus:ring-primary-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center dark:bg-primary-600 dark:hover:bg-primary-700 dark:focus:ring-primary-800"
                  to={"/library/books/add/upload"}
                  end
                >
                  Import a List of Books (.csv)
                </NavLink>
                <NavLink
                  className="block w-full text-white hover:text-white bg-majorelle hover:bg-hepatica focus:ring-4 focus:outline-none focus:ring-primary-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center"
                  to={"/library/books/add/manual"}
                  end
                >
                  Manually Add a Book
                </NavLink>
            </div>
          </div>
          {/* <h3>Snackbar poc</h3> */}
          {/* <button
            className="block w-full mb-2 bg-green-600 hover:bg-green-800 text-white font-bold py-2 px-4 rounded"
            onClick={() => handleShowSnackbar('added', 'Added book to your library')}>Add book - Toggle snackbar</button>
          <button
            className="block w-full mb-2 bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
            onClick={() => handleShowSnackbar('updated', 'Updated book meta information')}>Update book - Toggle snackbar</button>
          <button
            className="block w-full mb-2 bg-slate-500 hover:bg-slate-700 text-white font-bold py-2 px-4 rounded"
            onClick={() => handleShowSnackbar('removed', 'Book removed from your library')}>Remove book - Toggle snackbar</button>
          <button
            className="block w-full mb-2 bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded"
            onClick={() => handleShowSnackbar('error', 'Error - please try again later')}>Error - Toggle snackbar</button> */}
        </div>
        {/* <Snackbar
          message={snackbarMsg}
          open={snackbarOpen}
          onClose={handleCloseSnackbar}
          variant={snackbarVariant}
        /> */}
      </div>
    </section>
  );
}

export default AddBookGateway;
