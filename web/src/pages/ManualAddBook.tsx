import { useLocation, useNavigate } from 'react-router-dom';
import { SubmitHandler } from 'react-hook-form';
import BookForm from '../components/BookForm/BookForm';
import PageWithErrorBoundary from '../components/ErrorMessages/PageWithErrorBoundary';
import useAddBook from '../hooks/useAddBook';
import useStore from '../store/useStore';
import { BookFormData, Book } from '../types/api';
import Loading from '../components/Loading/Loading';

function ManualAddBook() {
  const { addBook, isLoading, refetchLibraryData } = useAddBook();
  const location = useLocation();
  const navigate = useNavigate();
  const { showSnackbar } = useStore();
  const bookData = location.state?.book || {};

  const handleAddBook: SubmitHandler<BookFormData> = async (data) => {
    const defaultDate = new Date().toISOString();

    const book: Book = {
      ...data,
      subtitle: data.subtitle || '',
      createdAt: defaultDate,
      lastUpdated: defaultDate,
      isbn10: data.isbn10 || '',
      isbn13: data.isbn13 || '',
      authors: data.authors
        .map((authorObj) => authorObj.author.trim())
        .filter((author) => author !== ''),
      genres: data.genres
        .map((genreObj) => genreObj.genre.trim())
        .filter((genre) => genre !== ''),
      tags: data.tags
        .map((tagObj) => tagObj.tag.trim())
        .filter((tag) => tag !== ''),
    };

    try {
      await addBook(book);
      await refetchLibraryData();
      showSnackbar('Book added successfully', 'added');
      navigate('/library');
    } catch (error) {
      console.error('Error adding book:', error);
      showSnackbar('Failed to add book. Please try again later.', 'error');
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

