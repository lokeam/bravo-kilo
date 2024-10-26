import { useCallback, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { addBook, updateBook, deleteBook, fetchUserBooks, fetchBooksAuthors, fetchBooksGenres, fetchBooksFormat, fetchBooksTags } from '../service/apiClient.service';
import { useUser } from './useUser';
import { Book, StringifiedBookFormData } from '../types/api';
import { invalidateLibraryQueries } from '../utils/invalidateQueries';

type BookOperation = 'add' | 'update' | 'delete';

const MAX_RETRIES = 3;
const RETRY_DELAY = 1000;

type BookOperationContext = {
  previousBook?: Book;
};

const useBookOperation = (operation: BookOperation, bookID?: string) => {
  const queryClient = useQueryClient();
  const { data: user } = useUser();
  const [isLoading, setIsLoading] = useState(false);

  const refetchLibraryData = useCallback(async () => {
    if (user?.id) {
      await Promise.all([
        queryClient.refetchQueries({ queryKey: ['books', user.id] }),
        queryClient.refetchQueries({ queryKey: ['booksAuthors', user.id] }),
        queryClient.refetchQueries({ queryKey: ['booksGenres', user.id] }),
        queryClient.refetchQueries({ queryKey: ['booksFormats', user.id] }),
        queryClient.refetchQueries({ queryKey: ['booksTags', user.id] }),
      ]);
    }
  }, [queryClient, user?.id]);

  // Prefetch library data
  const prefetchData = useCallback(() => {
    if (user?.id) {
      // Prefetch library data before navigation
      queryClient.prefetchQuery({
        queryKey: ['books', user.id],
        queryFn: () => fetchUserBooks(user.id),
      });
      queryClient.prefetchQuery({
        queryKey: ['booksAuthors', user.id],
        queryFn: () => fetchBooksAuthors(user.id),
      });
      queryClient.prefetchQuery({
        queryKey: ['booksGenres', user.id],
        queryFn: () => fetchBooksGenres(user.id),
      });
      queryClient.prefetchQuery({
        queryKey: ['booksFormats', user.id],
        queryFn: () => fetchBooksFormat(user.id),
      });
      queryClient.prefetchQuery({
        queryKey: ['booksTags', user.id],
        queryFn: () => fetchBooksTags(user.id),
      });
    }
  }, [queryClient, user?.id]);

  // Add this function above or below mutationFn
  const transformBookToFormData = (book: Book): StringifiedBookFormData => ({
    ...book,
    description: typeof book.description === 'string' ?
     book.description :
     JSON.stringify(book.description),
    notes: book.notes ? (
      typeof book.notes === 'string' ?
        book.notes : JSON.stringify(book.notes)
    ) : null,
    authors: book.authors.map(
      author => typeof author === 'string' ?
       { author } : author
      ),
    genres: book.genres.map(
      genre => typeof genre === 'string' ?
       { genre } : genre
      ),
    tags: book.tags?.map(
      tag => typeof tag === 'object' &&
      tag !== null &&
      'tag' in tag ?
        tag :
        { tag: tag as string }) ||
      [],
    formats: book.formats,
    pageCount: Number(book.pageCount),
    publishDate: book.publishDate || '',
    isbn10: book.isbn10 || '',
    isbn13: book.isbn13 || '',
    imageLink: book.imageLink || '', // Ensure imageLink is always a string
  });

  const mutationFn = (book: Book | StringifiedBookFormData | string) => {
    switch (operation) {
      case 'add':
        return addBook(book as StringifiedBookFormData);
      case 'update':
        return updateBook(book as Book, bookID!);
      case 'delete':
        return deleteBook(book as string);
    }
  };

  const mutation = useMutation<Book, Error, StringifiedBookFormData | string, BookOperationContext>({
    mutationFn,
    retry: (failureCount, error) => {
      if (failureCount < MAX_RETRIES) {
        console.log('Retrying mutation...');
        return true;
      }
      console.log('Mutation failed. Error:', error);
      return false;
    },
    retryDelay: RETRY_DELAY,
    onMutate: async (newBook) => {
      setIsLoading(true);
      if (operation === 'update' && bookID) {
        await queryClient.cancelQueries({ queryKey: ['book', bookID] });
        const previousBook = queryClient.getQueryData<Book>(['book', bookID]);
        if (typeof newBook !== 'string') {
          queryClient.setQueryData<Book>(['book', bookID], newBook as unknown as Book);
        }
        return { previousBook };
      }
      return {};
    },
    onSuccess: async (result) => {
      if (user?.id) {
        invalidateLibraryQueries(queryClient, user.id);
        await refetchLibraryData();
      }
      if (operation === 'update' && bookID) {
        queryClient.setQueryData<Book>(['book', bookID], result as Book);
      }
      prefetchData();
    },
    onError: (error, _, context) => {
      if (operation === 'update' && bookID && context?.previousBook) {
        queryClient.setQueryData<Book>(['book', bookID], context.previousBook);
      }
      console.error(`Error ${operation} book:`, error);
    },
    onSettled: () => {
      setIsLoading(false);
      if (bookID) {
        queryClient.invalidateQueries({ queryKey: ['book', bookID] });
      }
    }
  });

  const performOperationWithLoading = async (book: Book | string) => {
    setIsLoading(true);
    try {
      if (operation === 'delete') {
        await mutation.mutateAsync(book as string);
      } else {
        const formData = transformBookToFormData(book as Book);
        await mutation.mutateAsync(formData as any); // Use 'any' to bypass TypeScript check
      }
    } finally {
      setIsLoading(false);
    }
  };

  return {
    performOperation: performOperationWithLoading,
    isLoading,
    refetchLibraryData,
  };
};

export default useBookOperation;