import { useMemo} from 'react';
import LibraryNav from '../components/LibraryNav/LibraryNav';
import CardList from '../components/CardList/CardList';
import CardListSortHeader from '../components/CardList/CardListSortHeader';
import Snackbar from '../components/Snackbar/Snackbar';
import Loading from '../components/Loading/Loading';
import { defaultBookGenres, isBookGenresData, GenreData } from '../types/api';
import { sortBooks } from '../utils/ui';
import { Book } from '../types/api';

import useStore from '../store/useStore';
import useLibraryData from '../hooks/useLibraryData';

function Library() {
  const {
    activeTab,
    snackbarMessage,
    snackbarOpen,
    snackbarVariant,
    sortCriteria,
    sortOrder,
    hideSnackbar,
  } = useStore();

  const {
    books,
    authors:
    bookAuthors,
    genres: bookGenres,
    formats: bookFormats,
    isLoading,
    isError
  } = useLibraryData();

  const sortedBooks = useMemo(() => {
    if (!books || books.length === 0) {
      return [];
    }

    let booksToSort: Book[] = [];

    // Determine which books to sort based on the activeTab
    switch (activeTab) {
      case 'Audiobooks':
        booksToSort = bookFormats?.audioBook || [];
        break;
      case 'eBooks':
        booksToSort = bookFormats?.eBook || [];
        break;
      case 'Printed Books':
        booksToSort = bookFormats?.physical || [];
        break;
      default:
        booksToSort = books;
    }

    // Handle book sorting via utility function
    return sortBooks(booksToSort, sortCriteria, sortOrder);
  }, [activeTab, books, bookFormats, sortCriteria, sortOrder]);

  const safeBookGenres = isBookGenresData(bookGenres) ? bookGenres : defaultBookGenres;
  const { allGenres, ...remainingGenres } = safeBookGenres;
  const genreBooks = remainingGenres as { [key: string]: GenreData };

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <div>Error loading books</div>;
  }

  return (
    <div className="bk_lib flex flex-col items-center place-content-around px-5 antialiased mdTablet:pr-5 mdTablet:ml-24 h-screen pt-28">
      <LibraryNav />
      <CardListSortHeader sortedBooksCount={sortedBooks.length} />

      {activeTab === 'Authors' && bookAuthors?.allAuthors.length > 0 ? (
        <CardList
          allAuthors={bookAuthors.allAuthors}
          authorBooks={bookAuthors}
        />
      ) : activeTab === 'Genres' && bookGenres?.allGenres.length > 0 ? (
        <CardList
          allGenres={allGenres}
          genreBooks={genreBooks}
        />
      ) : (
        sortedBooks && sortedBooks.length > 0 &&
        <CardList
          books={sortedBooks}
          isSearchPage={false}
        />
      )}

      <Snackbar
        message={snackbarMessage || ''}
        open={snackbarOpen}
        variant={snackbarVariant || 'added'}
        onClose={hideSnackbar}
      />
    </div>
  )
}

export default Library;
