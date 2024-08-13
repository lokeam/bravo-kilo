import { useNavigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updateBook } from './../service/apiClient.service';
import useStore from '../store/useStore';
import { Book } from '../pages/Library';

const useUpdateBook = (bookID: string) => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { showSnackbar } = useStore();

  return useMutation({
    mutationFn: (book: Book) => updateBook(book, bookID),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['book', bookID] });
      queryClient.invalidateQueries({ queryKey: ['userBooks'] });
      queryClient.invalidateQueries({ queryKey: ['booksFormat'] });
      queryClient.invalidateQueries({ queryKey: ['userBookAuthors'] });
      queryClient.invalidateQueries({ queryKey: ['userBookGenres'] });
      showSnackbar('Book updated successfully', 'updated');
      navigate(`/library`);
    },
    onError: (error: any) => {
      console.log('useUpdateBook - Error updating book: ', error);
      showSnackbar(`Error updating book ${error.message}`, error);
    }
  })
}


export default useUpdateBook;