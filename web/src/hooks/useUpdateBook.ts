import { useNavigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updateBook } from './../service/apiClient.service';
import { Book } from '../pages/Library';

const useUpdateBook = (bookID: string) => {
  const navigate = useNavigate();

  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (book: Book) => updateBook(book, bookID),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['book', bookID] });
      navigate(`/library/books/${bookID}`);
    },
    onError: (error: any) => {
      console.log('useUpdateBook - Error updating book: ', error);
    }
  })
}


export default useUpdateBook;