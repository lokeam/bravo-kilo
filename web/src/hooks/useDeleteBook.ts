import { useNavigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { deleteBook } from '../service/apiClient.service';
import useStore from '../store/useStore';

const useDeleteBook = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { showSnackbar } = useStore();

  return useMutation({
    mutationFn: (bookID: string) => {
      console.log(`useDeleteBook - mutationFn called with bookID:`, bookID);
      return deleteBook(bookID);  // Ensure deleteBook expects bookID as string
    },
    onSuccess: () => {
      console.log(`useDeleteBook - onSuccess`);
      // Invalidate related queries to update the UI
      queryClient.invalidateQueries({ queryKey: ['book'] });
      queryClient.invalidateQueries({ queryKey: ['userBooks'] });
      queryClient.invalidateQueries({ queryKey: ['booksFormat'] });
      queryClient.invalidateQueries({ queryKey: ['userBookAuthors'] });
      queryClient.invalidateQueries({ queryKey: ['userBookGenres'] });
      showSnackbar('Book removed from library', 'removed');
      navigate('/library/');
    },
    onError: (error: any) => {
      console.log('useDeleteBook - Error deleting book: ', error);
      showSnackbar(`Error deleting book ${error.message}`, error);
    }
  });
};

export default useDeleteBook;
