import { useMemo } from 'react';
import Bookshelf from '../components/Bookshelf/Bookshelf';
import BarChartCard from '../components/Statistics/BarChartCard';
import TableCard from '../components/Statistics/TableCard';
import DonutChartCard from '../components/Statistics/DonutChartCard';
import EmptyHomeCard from '../components/ErrorMessages/EmptyHomeCard';
import Loading from '../components/Loading/Loading';
import PageWithErrorBoundary from '../components/ErrorMessages/PageWithErrorBoundary';

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
  //const isInitialLoad = !data || isLoading;
  // Check if all data arrays are empty
  const isEmpty = useMemo(() =>
    books.length === 0 &&
    booksByFormat.every(format => format.count === 0) &&
    booksByLang.length === 0 &&
    booksByGenre.length === 0 &&
    userTags.length === 0,
  [books, booksByFormat, booksByLang, booksByGenre, userTags]);

  if (isLoading) return (
    <div className="bk_home flex flex-col items-center px-5 antialiased mdTablet:pl-1 pr-5 mdTablet:ml-24 h-screen pt-12">
      <Loading />
    </div>
  );

  if (!isLoading && isEmpty) {
    return (
      <div className="bk_home flex flex-col items-center px-5 antialiased mdTablet:pl-1 pr-5 mdTablet:ml-24 h-screen pt-12">
        <EmptyHomeCard />
      </div>
    )
  }

  console.log('data package: ', data);
  console.log('booksByFormat: ', data);

  return (
    <PageWithErrorBoundary fallbackMessage="Error loading home page">
      <div className="bk_home bg-white-smoke flex flex-col items-center px-5 antialiased mdTablet:px-5 mdTablet:ml-24 h-screen pt-12 dark:bg-black">
        <div className="pb-20 mdTablet:pb-4 flex flex-col relative w-full max-w-7xl">
          <Bookshelf
            category="Recently updated"
            books={books || []}
            isLoading={isLoading}
          />
          <h2 className="text-left text-charcoal text-2xl font-bold inline-block max-w-full overflow-hidden text-ellipsis whitespace-nowrap select-none mb-4 dark:text-white">Statistics</h2>
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
    </PageWithErrorBoundary>
  )
}

export default Home;
