import { useMemo } from 'react';
import { useUser } from './useUser';
import useFetchData from './useFetchData';
import { fetchUserBooks, fetchBooksAuthors, fetchBooksGenres, fetchBooksFormat, fetchBooksTags } from '../service/apiClient.service';
import { Book, BookAuthorsData, BookGenresData, BookFormatData, BookTagsData } from '../types/api';


const useLibraryData = () => {
  const { data: user, isLoading: isUserLoading, isAuthenticated } = useUser();
  const userID = user?.id;
  const enabled = !!userID && isAuthenticated;

  const queryOptions = {
    enabled,
    staleTime: 0,
  };
  // Fetch books
  const {
    data: books,
    isLoading: isLoadingBooks,
    error: errorBooks
  } = useFetchData<Book[], number>({
    queryKey: ['books', userID],
    fetchFunction: fetchUserBooks,
    query: userID,
    ...queryOptions
  });

  const {
    data: authors,
    isLoading: isLoadingAuthors,
    error: errorAuthors
  } = useFetchData<BookAuthorsData, number>({
    queryKey: ['booksAuthors', userID],
    fetchFunction: fetchBooksAuthors,
    query: userID,
    ...queryOptions,
  });

  const {
    data: genres,
    isLoading: isLoadingGenres,
    error: errorGenres
  } = useFetchData<BookGenresData, number>({
    queryKey: ['booksGenres', userID],
    fetchFunction: fetchBooksGenres,
    query: userID,
    ...queryOptions,
  });

  const {
    data: formats,
    isLoading: isLoadingFormats,
    error: errorFormats
  } = useFetchData<BookFormatData, number>({
    queryKey: ['booksFormats', userID],
    fetchFunction: fetchBooksFormat,
    query: userID,
    ...queryOptions,
  });

  const {
    data: tags,
    isLoading: isLoadingTags,
    error: errorTags
  } = useFetchData<BookTagsData, number>({
    queryKey: ['booksTags', userID],
    fetchFunction: fetchBooksTags,
    query: userID,
    ...queryOptions,
  });

  const isLoading = isUserLoading || isLoadingBooks || isLoadingAuthors || isLoadingGenres || isLoadingFormats || isLoadingTags;

  const error = useMemo(() => {
    if (errorBooks) return errorBooks;
    if (errorAuthors) return errorAuthors;
    if (errorGenres) return errorGenres;
    if (errorFormats) return errorFormats;
    if (errorTags) return errorTags;
    return null;
  }, [errorBooks, errorAuthors, errorGenres, errorFormats, errorTags]);

  return {
    user,
    books,
    authors,
    genres,
    formats,
    tags,
    isLoading,
    isError: !!error,
    error,
    isAuthenticated,
  };
};

export default useLibraryData;

