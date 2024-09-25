import { useMemo } from 'react';
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
import useLibraryData from '../hooks/useLibraryData';


function Library() {
  const {
    activeTab,
    sortCriteria,
    sortOrder,
  } = useStore();

  const {
    books,
    authors:
    bookAuthors,
    genres: bookGenres,
    formats: bookFormats,
    tags: bookTags,
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

  const safeBookTags = isBookTagsData(bookTags) ?  bookTags : defaultBookTags;
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

  if (isLoading) {
    return (
      <div className="bk_lib flex flex-col items-center px-5 pt-12 antialiased mdTablet:pl-1 pr-5 mdTablet:ml-24 h-screen">
        <Loading />
      </div>
    );
  }

  console.log('bookTags: ', bookTags);

  return (
    <PageWithErrorBoundary fallbackMessage="Error loading library">
      <div className="bk_lib min-h-screen bg-white bg-cover flex flex-col items-center px-5 antialiased mdTablet:pl-1 pr-5 pt-5 mdTablet:ml-24 dark:bg-black">
        { isEmptyLibrary ?
          <EmptyLibraryCard /> :
          (
            <>
              <LibraryNav />
              <CardListSortHeader sortedBooksCount={sortedBooks.length} />
              {renderCardList()}
            </>
          )
        }
      </div>
    </PageWithErrorBoundary>
  )
}

export default Library;
