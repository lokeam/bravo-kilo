import useBookOperation from './useBookOperation';
import { StringifiedBookFormData } from '../types/api';

const useAddBook = () => {
  const { performOperation, isLoading, refetchLibraryData } = useBookOperation('add');

  return {
    addBook: (bookData: StringifiedBookFormData) => performOperation(bookData),
    isLoading,
    refetchLibraryData,
  };
};

export default useAddBook;
