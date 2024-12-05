import useBookOperation from './useBookOperation';
import { StringifiedBookFormData } from '../types/api';

const useUpdateBook = (bookID: string) => {
  const { performOperation, isLoading, refetchLibraryData } = useBookOperation('update', bookID);

  return {
    updateBook: (book: StringifiedBookFormData) => performOperation(book),
    isLoading,
    refetchLibraryData,
  };
};

export default useUpdateBook;
