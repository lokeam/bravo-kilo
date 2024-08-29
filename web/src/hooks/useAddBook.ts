import { useNavigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { addBook } from '../service/apiClient.service';
import useStore from '../store/useStore';
import { Book } from '../types/api';

const useAddBook = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { showSnackbar } = useStore();

  return useMutation({
    mutationFn: (book: Book) => {
      console.log(`useAddBook - mutationFn called with book:`, book); // Add this line
      return addBook(book);
    },
    onSuccess: () => {
      console.log(`useAddBook - onSuccess`); // Add this lineuserBooks
      queryClient.invalidateQueries({ queryKey: ['book'] });
      queryClient.invalidateQueries({ queryKey: ['userBooks'] });
      queryClient.invalidateQueries({ queryKey: ['booksFormat'] });
      queryClient.invalidateQueries({ queryKey: ['userBookAuthors'] });
      queryClient.invalidateQueries({ queryKey: ['userBookGenres'] });
      showSnackbar('Book added', 'added');
      navigate('/library/');
    },
    onError: (error: any) => {
      console.log('useAddBook - Error updating book: ', error);
      showSnackbar(`Error updating book ${error.message}`, error);
    }
  });
};

export default useAddBook;
