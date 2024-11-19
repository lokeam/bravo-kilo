import { useMutation, useQueryClient } from '@tanstack/react-query';
import { BookAPIPayload } from '../../../types/api';
import { queryKeys } from '../../constants/queryKeys';
import {
  addBook,
  updateBook,
  deleteBook,
} from '../../../service/apiClient.service';

export function useAddBook() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (book: BookAPIPayload) => addBook(book),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.library.root });
    },
  });
}

export function useUpdateBook() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ book, bookId }: { book: BookAPIPayload; bookId: string }) =>
      updateBook(book, bookId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.library.root });
    },
  });
}