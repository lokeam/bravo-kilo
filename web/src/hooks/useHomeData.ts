import { useCallback } from 'react';
import { useUser } from './useUser';
import useFetchBooks from './useFetchBooks';
import useFetchBooksFormat from './useFetchBooksFormat';
import useFetchHomepageData from './useFetchHomepageData';
import { AggregatedHomePageData } from '../types/api';

const useHomePageData = () => {
  const { data: user, isLoading: isUserLoading, isAuthenticated } = useUser();
  const userId = user?.id;
  const { data: books, isLoading: isLoadingBooks, isError: isErrorBooks } = useFetchBooks(userId, !!userId && isAuthenticated);
  const { data: booksByFormat, isLoading: isLoadingFormat, isError: isErrorFormat } = useFetchBooksFormat(userId, !!userId && isAuthenticated);
  const { data: homepageStats, isLoading: isLoadingStats, isError: isErrorStats } = useFetchHomepageData(userId, !!userId && isAuthenticated);

  const isLoading = isUserLoading && (isLoadingBooks || isLoadingFormat || isLoadingStats);
  const error = isErrorBooks || isErrorFormat || isErrorStats
    ? new Error('Error fetching data')
    : null;

  const data = useCallback((): AggregatedHomePageData | null => {
    if (books && booksByFormat && homepageStats) {
      return { books, booksByFormat, homepageStats };
    }
    return null;
  }, [books, booksByFormat, homepageStats]);

  return { data: data(), isLoading, error };
};

export default useHomePageData;
