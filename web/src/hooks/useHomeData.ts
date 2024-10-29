import { useMemo } from 'react';
import { useUser } from './useUser';
import useFetchBooks from './useFetchBooks';
import useFetchBooksFormat from './useFetchBooksFormat';
import useFetchHomepageData from './useFetchHomepageData';
import { AggregatedHomePageData, BooksByFormat } from '../types/api';

interface HomePageStats {
  authorsList: Array<{ name: string; count: number }>;
  booksByGenre: Array<{ genre: string; count: number }>;
  booksByLanguage: Array<{ language: string; count: number }>;
  userTags: Array<{ label: string; count: number }>;
}

interface FormatStats {
  formats: Array<{ format: string; count: number }>;
}

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

      // Ensure we have valid arrays for all stats
      const safeHomepageStats = {
        userAuthors: {
          booksByAuthor: homepageStats.userAuthors?.booksByAuthor || []
        },
        userBkGenres: {
          booksByGenre: homepageStats.userBkGenres?.booksByGenre || []
        },
        userBkLang: {
          booksByLang: homepageStats.userBkLang?.booksByLang || []
        },
        userTags: {
          userTags: homepageStats.userTags?.userTags || []
        }
      };

      return {
        books,
        booksByFormat: formattedBooksByFormat,
        homepageStats: safeHomepageStats
      };
    }
    return null;
  }, [books, booksByFormat, homepageStats]);

  return { data, isLoading, error };
};

export default useHomePageData;