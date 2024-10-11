import useBookOperation from './useBookOperation';

const useDeleteBook = () => {
  const { performOperation, isLoading, LoadingComponent } = useBookOperation('delete');

  return {
    deleteBook: (bookID: string) => performOperation(bookID),
    isLoading,
    LoadingComponent,
  };
};

export default useDeleteBook;
