
import { QueryClient } from '@tanstack/react-query';

export const invalidateLibraryQueries = (
  queryClient: QueryClient,
  userID: number,
  refetchType: 'active' | 'inactive' | 'all' | 'none' = 'active') => {
  const libraryQueryPrefixes = [
    'books',
    'booksAuthors',
    'booksGenres',
    'booksFormats',
    'booksTags',
    'booksHomepage',
  ];

  queryClient.invalidateQueries({
    predicate: (query) => {
      const queryKey = query.queryKey;
      return (
        Array.isArray(queryKey) &&
        queryKey.length === 2 &&
        libraryQueryPrefixes.includes(queryKey[0] as string) &&
        queryKey[1] === userID
      );
    },
    refetchType,
  });
};
