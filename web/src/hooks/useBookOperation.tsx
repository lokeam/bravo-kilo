import { useCallback, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { addBook, updateBook, deleteBook } from '../service/apiClient.service';
import { useUser } from './useUser';
import { Book, BookAPIPayload, StringifiedBookFormData, isStringifiedBookFormData } from '../types/api';
import { invalidateLibraryQueries } from '../utils/invalidateQueries';
import { getLibraryPageQueryKey } from '../queries/hooks/pages/library/useGetLibraryPageData';

type BookOperation = 'add' | 'update' | 'delete';

type BookOperationContext = {
  previousBook?: Book;
};

const MAX_RETRIES = 3;
const RETRY_DELAY = 1000;

type OperationPayload<T extends BookOperation> = T extends 'delete'
  ? string
  : StringifiedBookFormData;
const useBookOperation = <T extends BookOperation>(operation: T, bookID?: string) => {
  const queryClient = useQueryClient();
  const { data: user } = useUser();
  const [isLoading, setIsLoading] = useState(false);

  // Refetch library data as soon as mutation is successful
  const refetchLibraryData = useCallback(async () => {
    if (user?.id) {
      await Promise.all([
        queryClient.refetchQueries({
          queryKey: getLibraryPageQueryKey('books', user.id)
        }),
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
  // const prefetchData = useCallback(() => {
  //   if (user?.id) {
  //     // Prefetch library data before navigation
  //     queryClient.prefetchQuery({
  //       queryKey: ['books', user.id],
  //       queryFn: () => fetchUserBooks(user.id),
  //     });
  //     queryClient.prefetchQuery({
  //       queryKey: ['booksAuthors', user.id],
  //       queryFn: () => fetchBooksAuthors(user.id),
  //     });
  //     queryClient.prefetchQuery({
  //       queryKey: ['booksGenres', user.id],
  //       queryFn: () => fetchBooksGenres(user.id),
  //     });
  //     queryClient.prefetchQuery({
  //       queryKey: ['booksFormats', user.id],
  //       queryFn: () => fetchBooksFormat(user.id),
  //     });
  //     queryClient.prefetchQuery({
  //       queryKey: ['booksTags', user.id],
  //       queryFn: () => fetchBooksTags(user.id),
  //     });
  //   }
  // }, [queryClient, user?.id]);

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
      default: {
        throw new Error(`Unsupported operation: ${operation}`);
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

        // Immediately invalidate all library-related queries
        await Promise.all([
          queryClient.invalidateQueries({
            queryKey: getLibraryPageQueryKey('books', user.id)
          }),
          // Existing invalidations
          invalidateLibraryQueries(queryClient, user.id)
        ])
      }
      if (operation === 'update' && bookID) {
        queryClient.setQueryData<Book>(['book', bookID], result as Book);
      }
      await refetchLibraryData();
      //prefetchData();
    },
    onError: (error, _, context) => {
      if (operation === 'update' && bookID && context?.previousBook) {
        queryClient.setQueryData<Book>(['book', bookID], context.previousBook);
      }
      console.error(`Error ${operation} book:`, error);
    },
    onSettled: async () => {
      setIsLoading(false);
      if (bookID) {
        queryClient.invalidateQueries({ queryKey: ['book', bookID] });
      }
      // Force a refetch of homepage data
      if (user?.id) {
        await Promise.all([
          queryClient.invalidateQueries({
            queryKey: ['booksHomepage', user.id]
          })
        ]);
        // Force immediate refetch
        await refetchLibraryData();
      }
    }
  });

  const performOperationWithLoading = async (payload: OperationPayload<T>) => {
    setIsLoading(true);
    try {
      if (operation === 'delete') {
        // For delete operations, we expect either a string ID or a Book with an ID
        if (typeof payload !== 'string') {
          throw new Error('Invalid book data format');
        }
        await mutation.mutateAsync(payload);
      } else {
        // For add/update operations, ensure we're using StringifiedBookFormData
        if (typeof payload === 'string') {
          throw new Error('Invalid book data format');
        }
        if (!isStringifiedBookFormData(payload)) {
          throw new Error('Invalid book data format');
        }
        await mutation.mutateAsync(payload);
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
