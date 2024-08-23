import useFetchBooks from '../hooks/useFetchBooks';
import useFetchBooksFormat from '../hooks/useFetchBooksFormat';
import useFetchHomepageData from '../hooks/useFetchHomepageData';
import { useAuth } from '../components/AuthContext';

import Bookshelf from '../components/Bookshelf/Bookshelf';
import BarChartCard from '../components/Statistics/BarChartCard';
import TableCard from '../components/Statistics/TableCard';
import DonutChartCard from '../components/Statistics/DonutChartCard';

const Home = () => {
  const { user } = useAuth();
  const userID = user?.id || 0;

  const { data: books, isLoading: isLoadingBooks, isError: isErrorBooks } = useFetchBooks(userID, true);
  const { data: booksFormat, isLoading: isLoadingFormat, isError: isErrorFormat } = useFetchBooksFormat(userID, true)
  const { data: booksHp, isLoading: isLoadingHp, isError: isErrorHp } = useFetchHomepageData(userID, true)
  const { audioBook = [], physical = [], eBook = [] } = booksFormat || {};

  const booksByLang = booksHp?.userBkLang?.booksByLang || [];
  const booksByGenre = booksHp?.userBkGenres?.booksByGenre || [];
  const totalBooks = books && books.length || 0;
  const userTags = booksHp?.userTags?.userTags || [];
  const bookFormats = [
    { label: "Physical", count: physical.length },
    { label: "eBook", count: eBook.length },
    { label: "Audio", count: audioBook.length },
  ];

  console.log('testing books: ', books);

  if (isLoadingBooks || isLoadingFormat || isLoadingHp ) return <div>Loading...</div>;
  if (isErrorBooks || isErrorFormat) return <div>Error loading data</div>;
  if (isErrorHp) return <div>Error loading hp data</div>;

  return (
    <div className="bk_home flex flex-col items-center px-5 antialiased mdTablet:px-1 mdTablet:ml-24 h-screen pt-28">

      <div className="pb-20 mdTablet:pb-4 flex flex-col relative w-full max-w-7xl">
        {/* ------------------- Recently Edited ------------------- */}
        <Bookshelf category="Recently updated" books={books || []} isLoading={isLoadingBooks} />

        {/* ------------------- Stats: Book Formats ------------------- */}
        <h2 className="text-left text-2xl font-bold text-white inline-block max-w-full overflow-hidden text-ellipsis whitespace-nowrap select-none mb-4">Statistics</h2>

        {/* ------------------- Stat Cards ------------------- */}
        <div className="grid grid-cols-12 gap-6">
          {/* Format data */}
          <DonutChartCard bookFormats={bookFormats}/>

          {/* Language data */}
          <BarChartCard booksByLang={booksByLang} totalBooks={totalBooks} />

          {/* Genre data */}
          <BarChartCard booksByGenre={booksByGenre} totalBooks={totalBooks}/>

          {/* Tag Data */}
          <TableCard userTags={userTags}/>
        </div>
      </div>
    </div>
  )
}

export default Home;
