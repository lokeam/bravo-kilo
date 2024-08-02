import useFetchBooks from "../hooks/useFetchBooks";
import useFetchBooksCount from "../hooks/useFetchBooksCount";
import { useLocation } from "react-router-dom";

import Bookshelf from "../components/Bookshelf/Bookshelf";

import { BsBookHalf } from "react-icons/bs";
import { LuFile } from "react-icons/lu";
import { LuFileAudio } from "react-icons/lu";

const Home = () => {
  const { search } = useLocation();
  const query = new URLSearchParams(search);
  const userID = parseInt(query.get('userID') || '0', 10);

  console.log('Home component - userID: ', userID);

  const { data: books, isLoading: isLoadingBooks, isError: isErrorBooks } = useFetchBooks(userID);
  const { data: booksCount, isLoading: isLoadingCount, isError: isErrorCount} = useFetchBooksCount(userID);

  console.log('dashboard home - books: ', books);
  console.log('dashboard home == booksCount: ', booksCount);

  if (isLoadingBooks || isLoadingCount) return <div>Loading...</div>;
  if (isErrorBooks || isErrorCount) return <div>Error loading data</div>;

  return (
    <div className="bk_home flex flex-col items-center px-5 antialiased md:px-1 md:ml-24 h-screen pt-40">

      <div className="pb-20 md:pb-4 flex flex-col relative w-full max-w-7xl mt-8">
        {/* ------------------- Recently Edited ------------------- */}
        <Bookshelf category="Recently updated" books={books || []} isLoading={isLoadingBooks} />

        {/* ------------------- Add Category Bookshelf ------------------- */}
        <button className="mb-8">Add a shelf</button>

        {/* ------------------- Stats: Book Formats ------------------- */}
        <h2 className="text-left text-2xl font-bold text-white inline-block max-w-full overflow-hidden text-ellipsis whitespace-nowrap select-none mb-4">Statistics</h2>

        <div className="stat_shelf__formats min-[817px]:grid grid-cols-3 min-[817px]:gap-5 xl:gap-16">

          <div className="stat_1 flex flex-col h-40 text-center justify-center items-center rounded bg-ebony-clay w-full p-12">
            <div className="flex flex-row items-baseline text-5xl font-bold leading-9 whitespace-nowrap mb-2">
              <BsBookHalf color={"#6366F1"} size={32} className="mr-4" />
              <span>{booksCount?.physical || 0}</span>
            </div>
            <div className="col-start-1 whitespace-nowrap text-xl leading-7 font-semibold mb-2">Total Physical Books</div>
          </div>

          <div className="stat_2 flex flex-col h-40 text-center justify-center items-center rounded bg-ebony-clay w-full p-12 mt-4 min-[817px]:mt-0">
            <div className="flex flex-row items-baseline text-5xl font-bold leading-9 whitespace-nowrap mb-2">
              <LuFile color={"#6366F1"} size={32} className="mr-4" />
              <span>{booksCount?.eBook || 0}</span>
            </div>
            <div className="col-start-1 whitespace-nowrap text-xl leading-7 font-semibold mb-2">Total eBooks</div>

          </div>

          <div className="stat_3 flex flex-col h-40 text-center justify-center items-center rounded bg-ebony-clay w-full p-12 mt-4 min-[817px]:mt-0">
            <div className="flex flex-row items-baseline text-5xl font-bold leading-9 whitespace-nowrap mb-2">
              <LuFileAudio color={"#6366F1"} size={32} className="mr-4" />
              <span>{booksCount?.audioBook || 0}</span>
            </div>
            <div className="col-start-1 whitespace-nowrap text-xl leading-7 font-semibold mb-2">Total Audio Books</div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Home;