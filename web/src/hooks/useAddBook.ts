import useBookOperation from './useBookOperation';
import { Book } from '../types/api';

const useAddBook = () => {
  const { performOperation, isLoading, refetchLibraryData } = useBookOperation('add');

  return {
    addBook: (book: Book) => performOperation(book),
    isLoading,
    refetchLibraryData,
  };
};

export default useAddBook;
