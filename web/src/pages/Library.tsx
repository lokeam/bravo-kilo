import { useMemo } from 'react';
import { useUser } from '../hooks/useUser';
import { useGetLibraryPageData } from '../queries/hooks/pages/library/useGetLibraryPageData';

import LibraryNav from '../components/LibraryNav/LibraryNav';
import CardList from '../components/CardList/CardList';
import CardListSortHeader from '../components/CardList/CardListSortHeader';
import Loading from '../components/Loading/Loading';
import EmptyLibraryCard from '../components/ErrorMessages/EmptyLibraryCard';
import PageWithErrorBoundary from '../components/ErrorMessages/PageWithErrorBoundary';
import { sortBooks } from '../utils/ui';
import { Book } from '../types/api';
import {
  defaultBookGenres,
  isBookGenresData,
  GenreData,
  defaultBookTags,
  isBookTagsData,
  TagData,
} from '../types/api';
import useStore from '../store/useStore';


function Library() {
  const { data: user, isLoading: isUserLoading } = useUser();
  const {
    activeTab,
    sortCriteria,
    sortOrder,
  } = useStore();

  // Early return if no user ID to satisfy TypeScript
  if (!user?.id) {
    return (
      <div className="bk_lib flex flex-col items-center px-5 pt-12 antialiased mdTablet:pl-1 pr-5 mdTablet:ml-24 h-screen">
        <Loading />
      </div>
    );
  }

  const {
    data: libraryData,
    isLoading: isLibraryLoading,
    error: libraryError
  } = useGetLibraryPageData({
    userID: user.id,
    domain: 'books'
  });

  // Destructure and provide type-safe defaults
  const {
    books = [],
    booksByAuthors: bookAuthors = { allAuthors: [] },
    booksByGenres: bookGenres = defaultBookGenres,
    booksByFormat: bookFormats = { audioBook: [], eBook: [], physical: [] },
    booksByTags: bookTags = defaultBookTags,
  } = libraryData?.data || {};

  const sortedBooks = useMemo(() => {
    if (!books || books.length === 0) {
      return [];
    }

    let booksToSort: Book[] = [];

    // Determine which books to sort based on the activeTab
    switch (activeTab) {
      case 'Audiobooks':
        booksToSort = Array.isArray(bookFormats?.audioBook) ? bookFormats.audioBook : [];
        break;
      case 'eBooks':
        booksToSort = Array.isArray(bookFormats?.eBook) ? bookFormats.eBook : [];
        break;
      case 'Printed Books':
        booksToSort = Array.isArray(bookFormats?.physical) ? bookFormats.physical : [];
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

  const safeBookTags = isBookTagsData(bookTags) ? bookTags : defaultBookTags;
  console.log('testing safeBookTags: ', safeBookTags);

  const { allTags, ...remainingTags } = safeBookTags;
  const tagBooks = remainingTags as { [key: string]: TagData };

  const renderCardList = () => {
    if (activeTab === 'Authors' && bookAuthors?.allAuthors?.length > 0) {
      return <CardList allAuthors={bookAuthors.allAuthors} authorBooks={bookAuthors} />;
    }
    if (activeTab === 'Genres' && bookGenres?.allGenres?.length > 0) {
      return <CardList allGenres={allGenres} genreBooks={genreBooks} />;
    }
    if (activeTab === 'Tags' && bookTags?.allTags?.length > 0) {
      return <CardList allTags={allTags} tagBooks={tagBooks} />;
    }
    if (sortedBooks?.length > 0) {
      return <CardList books={sortedBooks} isSearchPage={false} />;
    }
    return null;
  };

  const isEmpty = (arr?: any[]) => !arr || arr.length < 1;

  const isEmptyLibrary = useMemo(() => {
    return (
      isEmpty(books) &&
      isEmpty(bookAuthors?.allAuthors) &&
      isEmpty(bookGenres?.allGenres) &&
      isEmpty(Object.keys(bookFormats || {}))
    );
  }, [books, bookAuthors, bookGenres, bookFormats]);

  const isLoading = isUserLoading || isLibraryLoading;
  if (isLoading) {
    return (
      <div className="bk_lib flex flex-col items-center px-5 pt-12 antialiased mdTablet:pl-1 pr-5 mdTablet:ml-24 h-screen">
        <Loading />
      </div>
    );
  }

  console.log('sortedBooks: ', sortedBooks);

  // Handle error states
  if (libraryError) {
    console.error('Library data fetch error:', {
      error: libraryError,
      userID: user?.id
    });

    return (
      <PageWithErrorBoundary fallbackMessage={libraryError.message || "Error loading library"}>
        <div className="bk_lib flex flex-col items-center px-5 pt-12">
          Error occurred while loading library data
        </div>
      </PageWithErrorBoundary>
    );
  }

  return (
    <PageWithErrorBoundary fallbackMessage="Error loading library">
      <div className="bk_lib min-h-screen bg-white bg-cover flex flex-col items-center px-5 antialiased mdTablet:pl-1 pr-5 pt-5 mdTablet:ml-24 dark:bg-black">
        {isEmptyLibrary ? (
          <EmptyLibraryCard />
        ) : (
          <>
            <LibraryNav />
            <CardListSortHeader sortedBooksCount={sortedBooks.length} />
            {renderCardList()}
          </>
        )}
      </div>
    </PageWithErrorBoundary>
  );
}

export default Library;
