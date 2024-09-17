import { useMemo } from 'react';
import Bookshelf from '../components/Bookshelf/Bookshelf';
import BarChartCard from '../components/Statistics/BarChartCard';
import TableCard from '../components/Statistics/TableCard';
import DonutChartCard from '../components/Statistics/DonutChartCard';
import EmptyHomeCard from '../components/ErrorMessages/EmptyHomeCard';
import Loading from '../components/Loading/Loading';
import useHomePageData from '../hooks/useHomeData';
import { AggregatedHomePageData } from '../types/api';

function Home() {
  const { data, isLoading, error } = useHomePageData();
  const { books, booksByFormat, totalBooks, booksByLang, booksByGenre, userTags } = useMemo(() => {
    const defaultData: AggregatedHomePageData = {
      books: [],
      booksByFormat: { audioBook: [], physical: [], eBook: [] },
      homepageStats: { userBkLang: { booksByLang: [] }, userBkGenres: { booksByGenre: [] }, userTags: { userTags: [] } }
    };

    if (!data) return {
      ...defaultData,
      totalBooks: 0,
      booksByFormat: [],
      booksByLang: [],
      booksByGenre: [],
      userTags: []
    };

    const { books = [], booksByFormat = { physical: [], eBook: [], audioBook: [] }, homepageStats = defaultData.homepageStats } = data;

    return {
      books,
      booksByFormat: [
        { label: "Physical", count: booksByFormat.physical?.length || 0 },
        { label: "eBook", count: booksByFormat.eBook?.length || 0 },
        { label: "Audio", count: booksByFormat.audioBook?.length || 0 },
      ],
      totalBooks: books.length || 0,
      booksByLang: homepageStats.userBkLang?.booksByLang || [],
      booksByGenre: homepageStats.userBkGenres?.booksByGenre || [],
      userTags: homepageStats.userTags?.userTags || [],
    };
  }, [data]);

  if (isLoading) return <Loading />;
  if (error) return <div>Error loading data: {error.message}</div>;

  // Check if all data arrays are empty
  const isEmpty = books.length === 0 &&
                  booksByFormat.every(format => format.count === 0) &&
                  booksByLang.length === 0 &&
                  booksByGenre.length === 0 &&
                  userTags.length === 0;

  if (isEmpty) {
    return <EmptyHomeCard />;
  }

  return (
    <div className="bk_home flex flex-col items-center px-5 antialiased mdTablet:pl-1 pr-5 mdTablet:ml-24 h-screen pt-28">

      <div className="pb-20 mdTablet:pb-4 flex flex-col relative w-full max-w-7xl">

        <Bookshelf
          category="Recently updated"
          books={books || []}
          isLoading={isLoading}
        />

        <h2 className="text-left text-2xl font-bold text-white inline-block max-w-full overflow-hidden text-ellipsis whitespace-nowrap select-none mb-4">Statistics</h2>

        <div className="grid grid-cols-12 gap-6">
          {/* Format data */}
          <DonutChartCard bookFormats={booksByFormat}/>

          {/* Language data */}
          <BarChartCard
            booksByLang={booksByLang}
            totalBooks={totalBooks}
          />

          {/* Genre data */}
          <BarChartCard
            booksByGenre={booksByGenre}
            totalBooks={totalBooks}
          />

          {/* Tag Data */}
          <TableCard userTags={userTags}/>
        </div>
      </div>
    </div>
  )
}

export default Home;
