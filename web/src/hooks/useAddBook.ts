import { useNavigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { addBook } from '../service/apiClient.service';
import { Book } from '../pages/Library';

const useAddBook = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (book: Book) => {
      console.log(`useAddBook - mutationFn called with book:`, book); // Add this line
      return addBook(book);
    },
    onSuccess: () => {
      console.log(`useAddBook - onSuccess`); // Add this line
      queryClient.invalidateQueries({ queryKey: ['book'] });
      navigate('/library/');
    },
    onError: (error: any) => {
      console.log('useAddBook - Error updating book: ', error);
    }
  });
};

export default useAddBook;
