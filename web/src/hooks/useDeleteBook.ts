import useBookOperation from './useBookOperation';

const useDeleteBook = () => {
  const { performOperation, isLoading, refetchLibraryData } = useBookOperation('delete');

  return {
    deleteBook: (bookID: string) => performOperation(bookID),
    isLoading,
    refetchLibraryData,
  };
};

export default useDeleteBook;
