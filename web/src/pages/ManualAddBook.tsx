import { useLocation, useNavigate } from 'react-router-dom';
import { SubmitHandler } from 'react-hook-form';
import BookForm from '../components/BookForm/BookForm';
import PageWithErrorBoundary from '../components/ErrorMessages/PageWithErrorBoundary';
import useAddBook from '../hooks/useAddBook';
import useStore from '../store/useStore';
import { BookFormData, StringifiedBookFormData, Book, isQuillDelta } from '../types/api';
import Loading from '../components/Loading/Loading';
import { transformFormData } from '../utils/bookFormHelpers';

function ManualAddBook() {
  const { addBook, isLoading, refetchLibraryData } = useAddBook();
  const location = useLocation();
  const navigate = useNavigate();
  const { showSnackbar } = useStore();
  const bookData = location.state?.book || {};

  // Update the type to match BookForm's output
  const handleAddBook: SubmitHandler<BookFormData> = async (formData) => {
    try {
      // Step 1: Validate form data
      if (!formData.title || !formData.description || !formData.language) {
        throw new Error('Required fields are missing');
      }

      // Step 2: Validate Quill content
      if (!isQuillDelta(formData.description)) {
        console.error('Invalid description format:', formData.description);
        throw new Error('Invalid description format');
      }

      if (formData.notes && !isQuillDelta(formData.notes)) {
        console.error('Invalid notes format:', formData.notes);
        throw new Error('Invalid notes format');
      }

      // Step 3: Transform form data to stringified format
      const stringifiedData = transformFormData(formData);

      // Step 4: Validate transformed data
      if (!Array.isArray(stringifiedData.authors) ||
          !Array.isArray(stringifiedData.genres) ||
          !Array.isArray(stringifiedData.tags)) {
        console.error('Invalid array fields:', {
          authors: stringifiedData.authors,
          genres: stringifiedData.genres,
          tags: stringifiedData.tags
        });
        throw new Error('Invalid data structure after transformation');
      }

      if (typeof stringifiedData.description !== 'string') {
        console.error('Invalid description after transformation:', stringifiedData.description);
        throw new Error('Invalid description after transformation');
      }

      // Step 5: Add metadata
      const enrichedData: StringifiedBookFormData = {
        ...stringifiedData,
      };

      console.log('Submitting book data:', enrichedData);

      // Step 6: Submit data
      await addBook(enrichedData);

      // Step 7: Handle success
      await refetchLibraryData();
      showSnackbar('Book added successfully', 'added');
      navigate('/library');

    } catch (error) {
      // Step 8: Handle errors
      console.error('Error in handleAddBook:', error);
      showSnackbar(
        error instanceof Error
          ? `Failed to add book: ${error.message}`
          : 'Failed to add book. Please try again later.',
        'error'
      );
    }
  };

  return (
    <PageWithErrorBoundary fallbackMessage="Error loading add manual page">
      <section className="addManual bg-white min-h-screen bg-cover relative flex flex-col items-center place-content-around px-5 antialiased mdTablet:pr-5 mdTablet:ml-24 dark:bg-black">
        <div className="text-left text-dark max-w-screen-mdTablet pb-24 md:pb-4 flex flex-col relative w-full">
          <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">Add Book</h2>
          { isLoading && <Loading /> }
          <BookForm
            onSubmit={handleAddBook}
            initialData={bookData}
            isLoading={isLoading}
          />
        </div>
      </section>
    </PageWithErrorBoundary>
  );
}

export default ManualAddBook;

