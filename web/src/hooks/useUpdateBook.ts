import useBookOperation from './useBookOperation';
import { Book } from '../types/api';

const useUpdateBook = (bookID: string) => {
  const { performOperation, isLoading } = useBookOperation('update', bookID);

  return {
    updateBook: (book: Book) => performOperation(book),
    isLoading,
  };
};

export default useUpdateBook;
