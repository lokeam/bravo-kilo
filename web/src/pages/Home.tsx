import { useEffect,useMemo } from 'react';
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
  const { data, isLoading } = useHomePageData();
  // Extract homepageStats for debugging
  const homepageStats = data?.homepageStats;
  useEffect(() => {
    if (homepageStats) {
      console.log('Homepage stats received:', homepageStats);
      console.log('Processed stats:', {
        authors: homepageStats.userAuthors?.booksByAuthor,
        genres: homepageStats.userBkGenres?.booksByGenre,
        languages: homepageStats.userBkLang?.booksByLang,
        tags: homepageStats.userTags?.userTags
      });
    }
  }, [homepageStats]);

  const { books, booksByFormat, totalBooks, booksByLang, booksByGenre, userTags, booksByAuthor } = useMemo(() => {
    const defaultData: AggregatedHomePageData = {
      books: [],
      booksByFormat: { audioBook: [], physical: [], eBook: [] },
      homepageStats: {
        userBkLang: { booksByLang: [] },
        userBkGenres: { booksByGenre: [] },
        userTags: { userTags: [] },
        userAuthors: { booksByAuthor: [] }
      }
    };

    if (!data) return {
      ...defaultData,
      totalBooks: 0,
      booksByFormat: [],
      booksByLang: [],
      booksByGenre: [],
      userTags: [],
      booksByAuthor: [],
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
      booksByAuthor: homepageStats.userAuthors?.booksByAuthor || [],
    };
  }, [data]);

  const isEmpty = useMemo(() => {
    const hasNoBooks = books.length === 0;
    const hasNoFormats = booksByFormat.every(format => format.count === 0);
    const hasNoLanguages = !booksByLang?.length;
    const hasNoGenres = !booksByGenre?.length;
    const hasNoAuthors = !booksByAuthor?.length;
    const hasNoTags = !userTags?.length;

    return hasNoBooks &&
           hasNoFormats &&
           hasNoLanguages &&
           hasNoGenres &&
           hasNoAuthors &&
           hasNoTags;
  }, [books, booksByFormat, booksByLang, booksByGenre, booksByAuthor, userTags]);

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

  console.log('===========');
  console.log('data package: ', data);
  console.log('booksByAuthor: ', booksByAuthor);
  console.log('books.length: ', books.length || 0);

  return (
    <PageWithErrorBoundary fallbackMessage="Error loading home page">
      <div className="bk_home bg-white-smoke dark:bg-dark-tone-ink flex flex-col items-center px-5 antialiased mdTablet:px-5 mdTablet:ml-24 h-screen pt-12">
        <div className="pb-20 mdTablet:pb-4 flex flex-col relative w-full max-w-7xl">
          <Bookshelf
            category="Recently updated"
            books={books || []}
            isLoading={isLoading}
          />
          <h2 className="text-left text-charcoal text-2xl font-bold inline-block max-w-full overflow-hidden text-ellipsis whitespace-nowrap select-none mb-4 dark:text-white">Statistics</h2>
          <div className="box-border grid grid-cols-12 gap-6">

            {/* Format data */}
            {booksByFormat && booksByFormat.length > 0 && (
              <DonutChartCard
                totalBooks={books.length || 0}
                bookFormats={booksByFormat}
              />
            )}

            {/* Tag Data */}
            {userTags && userTags.length > 0 && (
              <TableCard userTags={userTags}/>
            )}

            {/* Author data */}
            {booksByAuthor && booksByAuthor.length > 0 && (
              <BarChartCard
                booksByAuthor={booksByAuthor}
                totalBooks={totalBooks}
              />
            )}

            {/* Genre data */}
            {booksByGenre && booksByGenre.length > 0 && (
              <BarChartCard
                booksByGenre={booksByGenre}
                totalBooks={totalBooks}
              />
            )}

            {/* Language data */}
            {booksByLang && booksByLang.length > 0 && (
              <BarChartCard
                booksByLang={booksByLang}
                totalBooks={totalBooks}
              />
            )}


          </div>
        </div>
      </div>
    </PageWithErrorBoundary>
  )
}

export default Home;
