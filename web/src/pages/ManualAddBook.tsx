import { useLocation } from 'react-router-dom';
import { SubmitHandler } from 'react-hook-form';
import BookForm from '../components/BookForm/BookForm';
import PageWithErrorBoundary from '../components/ErrorMessages/PageWithErrorBoundary';
import useAddBook from '../hooks/useAddBook';
import { BookFormData, Book } from '../types/api';

function ManualAddBook() {
  const { mutate: addBook } = useAddBook();
  const location = useLocation();
  const bookData = location.state?.book || {};

  const handleAddBook: SubmitHandler<BookFormData> = (data) => {
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

    addBook(book), {
      onError: (error: Error) => {
        console.error('Error adding book:', error);
      },
    };
  };

  return (
    <PageWithErrorBoundary fallbackMessage="Error loading add manual page">
      <section className="addManual bg-white min-h-screen bg-cover relative flex flex-col items-center place-content-around px-5 antialiased mdTablet:pr-5 mdTablet:ml-24 dark:bg-black">
        <div className="text-left text-dark max-w-screen-mdTablet pb-24 md:pb-4 flex flex-col relative w-full">
          <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">Add Book</h2>
          <BookForm
            onSubmit={handleAddBook}
            initialData={bookData}
          />
        </div>
      </section>
    </PageWithErrorBoundary>
  );
}

export default ManualAddBook;

