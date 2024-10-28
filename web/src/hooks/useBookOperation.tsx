import { useCallback, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { addBook, updateBook, deleteBook, fetchUserBooks, fetchBooksAuthors, fetchBooksGenres, fetchBooksFormat, fetchBooksTags } from '../service/apiClient.service';
import { useUser } from './useUser';
import { Book, BookFormData, StringifiedBookFormData, isStringifiedBookFormData } from '../types/api';
import { invalidateLibraryQueries } from '../utils/invalidateQueries';
import Delta from 'quill-delta';
import { QuillContent } from '../types/api';

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

  /****** Utility Functions */
  const quillContentToString = (content: string | QuillContent | null |undefined): string => {
    if (typeof content === 'string') return content;
    if (!content) return '';
    if (content instanceof Delta) return JSON.stringify(content);
    if ('ops' in content) return JSON.stringify(content);
    return JSON.stringify(content);
  };

  const mutationFn = (book: StringifiedBookFormData | string) => {
    switch (operation) {
      case 'add':
        return addBook(book as StringifiedBookFormData);
      case 'update':
        if (typeof book === 'string') {
          throw new Error('Invalid book data for update operation');
        }
        // Convert Book to StringifiedBookFormData
        const bookForUpdate: StringifiedBookFormData = {
          ...book,
          publishDate: book.publishDate || '', // Ensure publishDate is always a string
          notes: book.notes ?? null, // Convert undefined to null
          formats: book.formats || [], // Ensure formats is always an array
          tags: book.tags || [], // Ensure tags is always an array
        };
        return updateBook(bookForUpdate, bookID!);
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

  const performOperationWithLoading = async (book: Book | StringifiedBookFormData) => {
    setIsLoading(true);
    try {
      if (operation === 'delete') {
        if (!('id' in book) || book.id === undefined) {
          console.error('Cannot delete a book without an id');
          // Todo: send to third party logging service
          return;
        }
        await mutation.mutateAsync(book.id.toString());
      } else {
      // For add/update operations, ensure we're using StringifiedBookFormData
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
