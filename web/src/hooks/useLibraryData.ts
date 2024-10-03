import { useQueries } from '@tanstack/react-query';
import { fetchUserBooks, fetchBooksAuthors, fetchBooksGenres, fetchBooksFormat, fetchBooksTags } from '../service/apiClient.service';
import { useUser } from './useUser';

const useLibraryData = () => {
  const { data: user, isLoading: isUserLoading, isAuthenticated } = useUser();
  const userID = user?.id;

  const queries = useQueries({
    queries: [
      {
        queryKey: ['books', userID],
        queryFn: () => fetchUserBooks(userID!),
        enabled: !!userID && isAuthenticated,
        staleTime: 1000 * 60 * 5,
      },
      {
        queryKey: ['booksAuthors', userID],
        queryFn: () => fetchBooksAuthors(userID!),
        enabled: !!userID && isAuthenticated,
        staleTime: 1000 * 60 * 5,
      },
      {
        queryKey: ['booksGenres', userID],
        queryFn: () => fetchBooksGenres(userID!),
        enabled: !!userID && isAuthenticated,
        staleTime: 1000 * 60 * 5,
      },
      {
        queryKey: ['booksFormats', userID],
        queryFn: () => fetchBooksFormat(userID!),
        enabled: !!userID && isAuthenticated,
        staleTime: 1000 * 60 * 5,
      },
      {
        queryKey: ['booksTags', userID],
        queryFn: () => fetchBooksTags(userID!),
        enabled: !!userID && isAuthenticated,
        staleTime: 1000 * 60 * 5,
      },
    ],
  });

  const [booksQuery, authorsQuery, genresQuery, formatsQuery, tagsQuery] = queries;

  return {
    user,
    books: booksQuery.data,
    authors: authorsQuery.data,
    genres: genresQuery.data,
    formats: formatsQuery.data,
    tags: tagsQuery.data,
    isLoading: isUserLoading || queries.some(query => query.isLoading),
    isError: queries.some(query => query.isError),
    error: queries.find(query => query.error)?.error,
    isAuthenticated,
  };
};

export default useLibraryData;
