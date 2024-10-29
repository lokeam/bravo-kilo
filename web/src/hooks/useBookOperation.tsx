import { useCallback, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { addBook, updateBook, deleteBook, fetchUserBooks, fetchBooksAuthors, fetchBooksGenres, fetchBooksFormat, fetchBooksTags } from '../service/apiClient.service';
import { useUser } from './useUser';
import { Book, BookAPIPayload, StringifiedBookFormData, isStringifiedBookFormData } from '../types/api';
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

  // Refetch library data as soon as mutation is successful
  const refetchLibraryData = useCallback(async () => {
    if (user?.id) {
      await Promise.all([
        queryClient.refetchQueries({ queryKey: ['books', user.id] }),
        queryClient.refetchQueries({ queryKey: ['booksAuthors', user.id] }),
        queryClient.refetchQueries({ queryKey: ['booksGenres', user.id] }),
        queryClient.refetchQueries({ queryKey: ['booksFormats', user.id] }),
        queryClient.refetchQueries({ queryKey: ['booksTags', user.id] }),
        queryClient.refetchQueries({ queryKey: ['booksHomepage', user.id] }),
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

  const mutationFn = (book: StringifiedBookFormData | string):Promise<Book> => {
    switch (operation) {
      case 'add': {
        if (typeof book === 'string') {
          throw new Error('Invalid book data for add operation');
        }
        const addPayload: BookAPIPayload = {
          ...book,
          description: { ops: JSON.parse(book.description).ops },
          notes: book.notes ? { ops: JSON.parse(book.notes).ops } : null,
        };
        return addBook(addPayload);
      }
      case 'update': {
        if (typeof book === 'string') {
          throw new Error('Invalid book data for update operation');
        }
        const updatePayload: BookAPIPayload = {
          ...book,
          publishDate: book.publishDate || '',
          formats: book.formats || [],
          tags: book.tags || [],
          description: JSON.parse(book.description),
          notes: book.notes ? JSON.parse(book.notes) : null,
        };
        return updateBook(updatePayload, bookID!);
      }
      case 'delete': {
        if (typeof book !== 'string') {
          throw new Error('Invalid book data for delete operation');
        }
        return deleteBook(book);
      }
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
      // Force a refetch of homepage data
      if (user?.id) {
        queryClient.invalidateQueries({ queryKey: ['booksHomepage', user.id] });
      }
    }
  });

  const performOperationWithLoading = async (book: Book | StringifiedBookFormData) => {
    setIsLoading(true);
    try {
      if (operation === 'delete') {
        // For delete operations, we expect either a string ID or a Book with an ID
        const bookId = typeof book === 'string'
          ? book
          : ('id' in book && book.id)
            ? book.id.toString()
            : null;
        if (!bookId) {
          throw new Error('Cannot delete a book without an id');
        }
        await mutation.mutateAsync(bookId);
      } else {
        // For add/update operations, ensure we're using StringifiedBookFormData
        if (typeof book === 'string') {
          throw new Error('Invalid book data format');
        }
        if (!isStringifiedBookFormData(book)) {
          throw new Error('Invalid book data format');
        }
        await mutation.mutateAsync(book);
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
