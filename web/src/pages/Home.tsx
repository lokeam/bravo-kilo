import useFetchBooks from '../hooks/useFetchBooks';
import useFetchBooksFormat from '../hooks/useFetchBooksFormat';
import { useAuth } from '../components/AuthContext';

import Bookshelf from '../components/Bookshelf/Bookshelf';
import LanguagesCard from '../components/Analytics/LanguagesCard';
import FormatsCard from '../components/Analytics/FormatsCard';
import TagsCard from '../components/Analytics/TagsCard';
import GenresCard from '../components/Analytics/GenresCard';


const Home = () => {
  const { user } = useAuth();
  const userID = user?.id || 0;

  console.log('Home component - userID: ', userID);

  const { data: books, isLoading: isLoadingBooks, isError: isErrorBooks } = useFetchBooks(userID, true);
  const { data: booksFormat, isLoading: isLoadingFormat, isError: isErrorFormat} = useFetchBooksFormat(userID, true)
  const { audioBook = [], physical = [], eBook = [] } = booksFormat || {};

  console.log('dashboard home - books: ', books);
  console.log('dashboard home == booksCount: ', booksFormat);

  if (isLoadingBooks || isLoadingFormat) return <div>Loading...</div>;
  if (isErrorBooks || isErrorFormat) return <div>Error loading data</div>;

  return (
    <div className="bk_home flex flex-col items-center px-5 antialiased mdTablet:px-1 mdTablet:ml-24 h-screen pt-28">

      <div className="pb-20 mdTablet:pb-4 flex flex-col relative w-full max-w-7xl">
        {/* ------------------- Recently Edited ------------------- */}
        <Bookshelf category="Recently updated" books={books || []} isLoading={isLoadingBooks} />

        {/* ------------------- Stats: Book Formats ------------------- */}
        <h2 className="text-left text-2xl font-bold text-white inline-block max-w-full overflow-hidden text-ellipsis whitespace-nowrap select-none mb-4">Statistics</h2>

        {/* <div className="stat_shelf__formats min-[817px]:grid grid-cols-3 min-[817px]:gap-5 xl:gap-16">

          <div className="stat_1 flex flex-col h-40 text-center justify-center items-center rounded bg-ebony-clay w-full p-12">
            <div className="flex flex-col items-center text-5xl font-bold leading-9 whitespace-nowrap mb-2">
              <BsBookHalf color={"red"} size={32} className="mb-2" />
              <span>{physical?.length || 0}</span>
            </div>
            <div className="col-start-1 whitespace-nowrap text-xl leading-7 font-semibold mb-2">Physical Books</div>
          </div>

          <div className="stat_2 flex flex-col h-40 text-center justify-center items-center rounded bg-ebony-clay w-full p-12 mt-4 min-[817px]:mt-0">
            <div className="flex flex-col items-center text-5xl font-bold leading-9 whitespace-nowrap mb-2">
              <LuFile color={"green"} size={32} className="mb-2" />
              <span>{eBook?.length || 0}</span>
            </div>
            <div className="col-start-1 whitespace-nowrap text-xl leading-7 font-semibold mb-2">eBooks</div>

          </div>

          <div className="stat_3 flex flex-col h-40 text-center justify-center items-center rounded bg-ebony-clay w-full p-12 mt-4 min-[817px]:mt-0">
            <div className="flex flex-col items-center text-5xl font-bold leading-9 whitespace-nowrap mb-2">
              <LuFileAudio color={"#352F99"} size={32} className="mb-2" />
              <span>{audioBook?.length || 0}</span>
            </div>
            <div className="col-start-1 whitespace-nowrap text-xl leading-7 font-semibold mb-2">Audio Books</div>
          </div>
        </div> */}

        {/* ------------------- Stat Cards ------------------- */}
        <div className="grid grid-cols-12 gap-6">

          {/* Formats data */}
          <div className="books_format_card_wrapper flex flex-col col-span-full lgMobile:col-span-6 mdTablet:col-span-4 bg-maastricht shadow-sm rounded-xl">
            <div className="books_format_header border-b border-gray-700/60 px-5 py-4">
              <h2 className="text-left text-lg font-semibold">Books By Format</h2>
            </div>
            <div className="books_format_card_header flex flex-wrap px-5 py-6">
              <div className="flex flex-col items-start min-w-[33%] py-2 mr-6">
                <div className="text-4xl text-left border-r border-gray-600 w-full font-bold text-gray-800 dark:text-gray-100 mr-2 mb-1">
                  <span>{physical?.length || 0}</span>
                </div>
                <div className="text-sm text-gray-500 dark:text-gray-400 text-upper font-semibold">Physical</div>
              </div>
              <div className="flex flex-col items-start min-w-[33%] pl-6 py-2 mr-6">
                <div className="text-4xl text-left border-r border-gray-600 w-full font-bold text-gray-800 dark:text-gray-100 mr-2 mb-1">
                  <span>{eBook?.length || 0}</span>
                </div>
                <div className="text-sm text-gray-500 dark:text-gray-400 text-upper font-semibold">eBooks</div>
              </div>
              <div className="flex flex-col items-start min-w-[33%] pl-6 py-2">
                <div className="text-4xl font-bold text-gray-800 dark:text-gray-100 mr-2 mb-1">
                  <span>{audioBook?.length || 0}</span>
                </div>
                <div className="text-sm text-gray-500 dark:text-gray-400 text-upper font-semibold">Audio</div>
              </div>
            </div>
            <div className="flex-grow pb-4">
              <FormatsCard />
            </div>
          </div>

          {/* Language data */}
          <LanguagesCard />

          {/* Genre data */}
          <GenresCard />

          {/* Tag Data */}
          <TagsCard />
        </div>
      </div>
    </div>
  )
}

export default Home;
