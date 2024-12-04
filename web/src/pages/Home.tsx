import { useEffect,useMemo } from 'react';
import Bookshelf from '../components/Bookshelf/Bookshelf';
import BarChartCard from '../components/Statistics/BarChartCard';
import TableCard from '../components/Statistics/TableCard';
import DonutChartCard from '../components/Statistics/DonutChartCard';
import EmptyHomeCard from '../components/ErrorMessages/EmptyHomeCard';
import Loading from '../components/Loading/Loading';
import PageWithErrorBoundary from '../components/ErrorMessages/PageWithErrorBoundary';

import { useUser } from '../hooks/useUser';
import { useGetHomePageData } from '../queries/hooks/pages/home/useGetHomePageData';
import {
  defaultHomePageData,
  defaultBookFormats,
  defaultHomePageStats,
  FormatCount,
  TransformedHomeData
} from '../types/api';
import { TAILWIND_HOMEPAGE_CLASSES } from '../consts/styleConsts';

function Home() {
  const { data: user, isLoading: isUserLoading } = useUser();

  // Early return if no user ID to satisfy TypeScript
  if (!user?.id) {
    return (
      <div className={TAILWIND_HOMEPAGE_CLASSES['LOADING_WRAPPER']}>
        <Loading />
      </div>
    );
  }

  const {
    data: homeData,
    isLoading: isHomeDataLoading,
    error: homeError
  } = useGetHomePageData({
    userID: user.id,
    domain: 'books'
  });


  // Extract homepageStats for debugging
  //const homepageStats = homeData?.homepageStats;

  useEffect(() => {
    if (homeData?.homepageStats) {  // Access nested data
      console.log('Homepage stats received:', homeData.homepageStats);
      console.log('Processed stats:', {
        authors: homeData.homepageStats.userAuthors?.booksByAuthor || [],
        genres: homeData.homepageStats.userBkGenres?.booksByGenre || [],
        languages: homeData.homepageStats.userBkLang?.booksByLang || [],
        tags: homeData.homepageStats.userTags?.userTags || []
      });
    }
  }, [homeData]);

  // Data transformation memoization
  const {
    books,
    booksByFormat,
    totalBooks,
    booksByLang,
    booksByGenre,
    userTags,
    booksByAuthor
  }: TransformedHomeData = useMemo(() => {
    if (!homeData) {
      return defaultHomePageData;
    }

    const {
      books = [],
      booksByFormat = defaultBookFormats,
      homepageStats = defaultHomePageStats,
    } = homeData;  // Ensure access from homeData.data

    // Transform booksByFormat into array of FormatCount
    const transformedBooksByFormat: FormatCount[] = [
      { label: 'Physical', count: booksByFormat.physical || 0 },
      { label: 'eBook', count: booksByFormat.eBook || 0 },
      { label: 'Audio', count: booksByFormat.audioBook || 0 },
    ];

    return {
      books,
      booksByFormat: transformedBooksByFormat,
      totalBooks: books.length || 0,
      booksByLang: homepageStats.userBkLang?.booksByLang || [],
      booksByGenre: homepageStats.userBkGenres?.booksByGenre || [],
      userTags: homepageStats.userTags?.userTags || [],
      booksByAuthor: homepageStats.userAuthors?.booksByAuthor || [],
    };
  }, [homeData]);

  // Empty state memoization
  const isEmpty = useMemo(() => {
    const hasNoBooks = books.length === 0;
    const hasNoFormats = (booksByFormat as FormatCount[]).every(
      (format: FormatCount) => format.count === 0
    );
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

  const isLoading = isUserLoading || isHomeDataLoading;

  if (isHomeDataLoading) return (
    <div className={TAILWIND_HOMEPAGE_CLASSES['LOADING_WRAPPER']}>
      <Loading />
    </div>
  );

  if (!isHomeDataLoading && isEmpty) {
    return (
      <div className={TAILWIND_HOMEPAGE_CLASSES['LOADING_WRAPPER']}>
        <EmptyHomeCard />
      </div>
    )
  }

  console.log('===========');
  console.log('data package: ', homeData);
  console.log('booksByAuthor: ', booksByAuthor);
  console.log('books.length: ', books.length || 0);


  // Handle error states
  if (homeError) {
    console.error('Home data fetch error: ', {
      error: homeError,
      userID: user?.id
    })
  }

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
