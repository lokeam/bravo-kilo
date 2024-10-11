import { useMemo } from 'react';
import { useUser } from './useUser';
import useFetchBooks from './useFetchBooks';
import useFetchBooksFormat from './useFetchBooksFormat';
import useFetchHomepageData from './useFetchHomepageData';
import { AggregatedHomePageData, BooksByFormat } from '../types/api';

const useHomePageData = () => {
  const { data: user, isLoading: isUserLoading, isAuthenticated } = useUser();
  const userId = user?.id;

  const enabled = !!userId && isAuthenticated;

  const {
    data: books,
    isLoading: isLoadingBooks,
    error: errorBooks
  } = useFetchBooks(userId, enabled);

  const {
    data: booksByFormat,
    isLoading: isLoadingFormat,
    error: errorFormat
  } = useFetchBooksFormat(userId, enabled);

  const {
    data: homepageStats,
    isLoading: isLoadingStats,
    error: errorStats
  } = useFetchHomepageData(userId, enabled);

  const isLoading = isUserLoading || isLoadingBooks || isLoadingFormat || isLoadingStats;

  const error = useMemo(() => {
    if (errorBooks) return errorBooks;
    if (errorFormat) return errorFormat;
    if (errorStats) return errorStats;
    return null;
  }, [errorBooks, errorFormat, errorStats]);

  const data: AggregatedHomePageData | null = useMemo(() => {
    if (books && homepageStats && booksByFormat) {
      const formattedBooksByFormat: BooksByFormat = {
        physical: Array.isArray(booksByFormat.physical) ? booksByFormat.physical : [],
        eBook: Array.isArray(booksByFormat.eBook) ? booksByFormat.eBook : [],
        audioBook: Array.isArray(booksByFormat.audioBook) ? booksByFormat.audioBook : [],
      };

      return { books, booksByFormat: formattedBooksByFormat, homepageStats };
    }
    return null;
  }, [books, booksByFormat, homepageStats]);

  return { data, isLoading, error };
};

export default useHomePageData;