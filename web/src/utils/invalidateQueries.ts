
import { QueryClient } from '@tanstack/react-query';

export const invalidateLibraryQueries = async (
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

  await Promise.all(
    libraryQueryPrefixes.map((prefix) => {
      queryClient.invalidateQueries({
        queryKey: [prefix, userID],
        refetchType,
        exact: true,
      });
    })
  );

  // Force an immediate refetch if refetchType isn't 'none'
  if (refetchType !== 'none') {
    await Promise.all(
      libraryQueryPrefixes.map(prefix => {
        queryClient.refetchQueries({
          queryKey: [prefix, userID],
          exact: true,
        })
      })
    )
  }
};
