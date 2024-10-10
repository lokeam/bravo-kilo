import { NavLink } from 'react-router-dom';
import { IoSearchOutline } from 'react-icons/io5';
import PageWithErrorBoundary from '../components/ErrorMessages/PageWithErrorBoundary';

const AddBookGateway = () => {

  return (
    <PageWithErrorBoundary fallbackMessage="Error loading home page">
      <section className="px-5 bg-white-smoke dark:bg-dark-tone-ink mdTablet:ml-24 h-full pt-12">
        <div className="flex flex-col items-center pb-8 mx-auto h-screen lg:py-0">
          <div className="w-full bg-white rounded-lg shadow-xl md:mt-0 sm:max-w-md xl:p-0 dark:bg-eight-ball dark:border dark:border-gray-700/60">
            <div className="p-6 space-y-4 md:space-y-6 sm:p-8">
              <h1 className="text-xl font-bold leading-tight tracking-tight text-gray-900 md:text-2xl dark:text-white">
                Add to Your Library
              </h1>
              <div className="space-y-4 md:space-y-6">
                <label htmlFor="topbar-search" className="sr-only">Search</label>
                <div className="flex flex-row relative mt-1">
                  <div className="flex absolute inset-y-0 left-0 items-center pl-3 pointer-events-none dark:text-white">
                    <IoSearchOutline
                      className="text-black dark:text-white"
                    />
                  </div>
                  <NavLink
                    className="bg-white dark:bg-dark-tone-ink text-charcoal border border-gray-600 dark:border-gray-700/60 hover:text-charcoal font-bold sm:text-sm rounded block w-full pl-9 p-2.5 placeholder:polo-blue placeholder:font-bold dark:text-az-white dark:hover:text-az-white"
                    to={"/library/books/search"}
                    end
                  >
                    Add book via Search
                  </NavLink>
                </div>
                  <div className="flex items-center justify-between">
                    <div className="h-0.5 w-full bg-gray-200 dark:bg-gray-600"></div>
                    <div className="text-center px-5 text-black dark:text-white">or</div>
                    <div className="h-0.5 w-full bg-gray-200 dark:bg-gray-600"></div>
                  </div>
                  <NavLink
                    className="block w-full text-white hover:text-white bg-vivid-blue hover:bg-vivid-blue-l focus:ring-4 focus:outline-none focus:ring-primary-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center dark:bg-vivid-blue dark:hover:bg-vivid-blue-d dark:focus:ring-primary-800 transition duration-500 ease-in-out"
                    to={"/library/books/add/upload"}
                    end
                  >
                    Import a List of Books (.csv)
                  </NavLink>
                  <NavLink
                    className="block w-full text-white hover:text-white bg-strong-violet hover:bg-strong-violet-l focus:ring-4 focus:outline-none focus:ring-primary-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center transition duration-500 ease-in-out"
                    to={"/library/books/add/manual"}
                    end
                  >
                    Manually Add a Book
                  </NavLink>
              </div>
            </div>
          </div>
        </div>
      </section>
    </PageWithErrorBoundary>
  );
}

export default AddBookGateway;
