import { useCallback, useEffect, useMemo, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useAuth } from '../components/AuthContext';
import useStore from '../store/useStore';
import useFetchBooks from '../hooks/useFetchBooks';
import CardList from '../components/CardList/CardList';
import Modal from '../components/Modal/Modal';
import Snackbar from '../components/Snackbar/Snackbar';

import '../components/Modal/Modal.css';
import { PiArrowsDownUp } from 'react-icons/pi';
import { fetchBooksAuthors, fetchBooksFormat, fetchBooksGenres } from '../service/apiClient.service';
import { Book, BookAuthorsData, BookGenresData, GenreData } from '../types/api';


// Type guard for Book Genre
function isBookGenresData(data: any): data is BookGenresData {
  return (
    data &&
    Array.isArray(data.allGenres) &&
    Object.values(data).some(
      (value) => {
        return (
          value !== null &&
          typeof value === 'object' &&
          'bookList' in value &&
          'genreImgs' in value
        );
      }
    )
  );
}

const Library = () => {
  // Zustand storage
  const {
    // Library Sorting
    sortCriteria, sortOrder, setSort, activeTab, setActiveTab, snackbarMessage, snackbarOpen, snackbarVariant, hideSnackbar,
  } = useStore();

  const [opened, setOpened] = useState(false);
  const { user } = useAuth();
  const userID = user && typeof user.id === 'string' ? parseInt(user.id, 10) : 0;
  const { data: books, isLoading, isError } = useFetchBooks(userID, true);

  const queryClient = useQueryClient();

  // Prefetch data for book formats
  useEffect(() => {
    queryClient.prefetchQuery({
      queryKey: ['booksFormat', userID],
      queryFn: () => fetchBooksFormat(userID)
    }).then(() => {
      console.log('Prefetched formats data is stored in cache:', queryClient.getQueryData(['booksFormat', userID]));
    });
  }, [userID, queryClient]);

  // Use useQuery to retrieve cached books authors
  console.log('Library page, userID: ', userID);
  const {
    data: bookAuthors = { allAuthors: [], },
    isLoading: isAuthorsLoading,
    isError: isAuthorsError,
  } = useQuery<BookAuthorsData>({
    queryKey: ['userBookAuthors', userID],
    queryFn: () => fetchBooksAuthors(userID),
    staleTime: Infinity,
    gcTime: 1000 * 60 * 60 * 24,
  });

  const defaultBookGenres: BookGenresData = {
    allGenres: [], // Correctly initialized as an array of strings
    placeholder: {
      bookList: [], // Matches `Book[]` type
      genreImgs: [], // Matches `string[]` type
    },
  };

  // Use useQuery to get cached book genres
  const {
    data: bookGenres = { allGenres: [], },
    isLoading: isGenresLoading,
    isError: isGenresError,
  } = useQuery<BookGenresData>({
    queryKey: ['userBookGenres', userID],
    queryFn: () => fetchBooksGenres(userID),
    staleTime: Infinity,
    gcTime: 1000 * 60 * 60 * 24,
  });

  // Retrieve cached books formats
  const bookFormats = queryClient.getQueryData<{
    audioBook: Book[],
    eBook: Book[],
    physical: Book[]
  }>(['booksFormat', userID]);



  const safeBookGenres = isBookGenresData(bookGenres)
  ? bookGenres
  : defaultBookGenres;

// Separate `allGenres` and retain the rest of the genres
console.log('Initial bookGenres:', bookGenres);

// Destructure `allGenres` from the rest of the data, keeping everything else intact
const { allGenres, ...remainingGenres } = safeBookGenres;

// Log `remainingGenres` before filtering to understand what's inside
console.log('Remaining genres before filtering:', remainingGenres);

// Directly use `remainingGenres` without further filtering
const genreBooks = remainingGenres as { [key: string]: GenreData };

// Log the final genreBooks object
console.log('Final genreBooks:', genreBooks);



  // Memoize book sorting
  const getSortedBooks = useCallback(
    (
      books: Book[],
      criteria: 'title' | 'publishDate' | 'author' | 'pageCount',
      order: 'asc' | 'desc'
    ) => {
      return books.slice().sort((a, b) => {
        switch (criteria) {
          case "title":
            return order === "asc" ? a.title.localeCompare(b.title) : b.title.localeCompare(a.title);

          case "publishDate": {
            const dateA = a.publishDate ? new Date(a.publishDate).getTime() : 0;
            const dateB = b.publishDate ? new Date(b.publishDate).getTime() : 0;
            return order === "asc" ? dateA - dateB : dateB - dateA;
          }

          case "author": {
            const aSurname = a.authors?.[0]?.split(" ").pop() || "";
            const bSurname = b.authors?.[0]?.split(" ").pop() || "";
            return order === "asc" ? aSurname.localeCompare(bSurname) : bSurname.localeCompare(aSurname);
          }

          default:
            return order === "asc" ? a.pageCount - b.pageCount : b.pageCount - a.pageCount;
        }
      });
    },
    []
  );


  const sortedBooks = useMemo(() => {
    if (!books || books.length === 0) {
      console.log('No books available');
      return [];
    }
    let booksToSort = [];
    if (activeTab === 'Audiobooks' && bookFormats) {
      console.log('Using audiobooks format');
      booksToSort = bookFormats.audioBook || [];
      console.log('set booksToSort to audioBooks: ', booksToSort);
    } else if (activeTab === 'eBooks' && bookFormats) {
      console.log('Using eBooks format');
      booksToSort = bookFormats.eBook || [];
      console.log('set booksToSort to ebooks: ', booksToSort);
    } else if (activeTab === 'Printed Books' && bookFormats) {
      console.log('Using printed books format');
      booksToSort = bookFormats.physical || [];
      console.log('set booksToSort to physicalBooks: ', booksToSort);
    } else {
      console.log('Using all books');
      booksToSort = books || [];
    }
    return getSortedBooks(booksToSort, sortCriteria, sortOrder);
  }, [activeTab, books, bookFormats, sortCriteria, sortOrder, getSortedBooks]);




  // Memoized Handlers
  const handleSort = useCallback(
    (criteria: "title" | "publishDate" | "author" | "pageCount") => {
      const order = sortOrder === 'asc' ? 'desc' : 'asc';
      setSort(criteria, order);
      setOpened(false);
    },
    [sortOrder, setSort]
  );

  const handleTabClick = useCallback(
    (tab: string) => {
      console.log(`Switching to tab: ${tab}`);
      setActiveTab(tab);
    },
    [setActiveTab]
  );

  if (isLoading || isAuthorsLoading || isGenresLoading) {
    return <div>Loading...</div>;
  }

  if (isError || isAuthorsError || isGenresError) {
    return <div>Error loading books</div>;
  }

  if (!bookAuthors || bookAuthors.allAuthors.length === 0) {
    console.log('No authors data available');
    return <div>No authors found</div>;
  }

  if (!bookGenres || bookGenres.allGenres.length === 0) {
    console.log('No genres data available');
    return <div>No genres found</div>;
  }

  const sortButtonTitle = {
    'title': 'Title: A to Z',
    'author': 'Author: A to Z',
    'publishDate': 'Release date: New to Old',
    'pageCount': 'Page count: Short to Long',
  };

  const openModal = () => setOpened(true);
  const closeModal = () => setOpened(false);


  console.log('Fetched books:', books);
  //console.log('Fetched authors:', bookAuthors);
  console.log('Fetched genres:', bookGenres);
  //console.log('Fetched formats:', bookFormats);
  console.log('-----');


  return (
    <div className="bk_lib flex flex-col items-center place-content-around px-5 antialiased mdTablet:pr-5 mdTablet:ml-24 h-screen pt-28">
      {/* Library Nav */}
      <div className="bookshelf_body relative w-full z-10 pb-8">
        <div className="bookshelf_grid_wrapper box-border ">
          <div className="bookshelf_grid_body box-content overflow-visible w-full">
            <ul className="bookshelf_grid_library text-left box-border grid grid-flow-col auto-cols-auto items-stretch gap-x-2.5 overflow-x-auto overflow-y-auto overscroll-x-none scroll-smooth snap-start snap-x snap-mandatory list-none m-0 pb-5">
              {['All', 'Audiobooks', 'eBooks', 'Printed Books', 'Authors', 'Genres'].map((tab) => (
                <li
                  key={tab}
                  className={`flex items-center text-nowrap cursor-pointer ${
                    activeTab === tab ? 'text-3xl font-bold text-white' : 'text-lg font-semibold text-cadet-gray'
                  }`}
                  onClick={() => handleTabClick(tab)}
                >
                  <span>{tab}</span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </div>

      {/* Libary Sort Modal  */}
      <Modal opened={opened} onClose={closeModal} title="">
        <button onClick={() => handleSort("publishDate")} className="flex flex-row bg-transparent mr-1">
          Release date: New to Old
        </button>
        <button onClick={() => handleSort("pageCount")} className="flex flex-row bg-transparent mr-1">
          Page count: Short to Long
        </button>
        <button onClick={() => handleSort("title")} className="flex flex-row bg-transparent mr-1">
          Title: A to Z
        </button>
        <button onClick={() => handleSort("author")} className="flex flex-row bg-transparent mr-1">
          Author: A to Z
        </button>
      </Modal>

      {/* Library Card List Header */}
      <div className="flex flex-row relative w-full max-w-7xl justify-between items-center text-left text-white border-b-2 border-solid border-zinc-700 pb-6 mb-2">
        {/* Number of total items in view  */}
        <div className="mt-1">{activeTab === 'Authors' ? bookAuthors?.allAuthors.length :sortedBooks?.length || 0} {activeTab === 'Authors' ? 'authors' : ' volumes'}</div>

        {/* Sort Button  */}
        <div className="flex flex-row">
          <button className="flex flex-row justify-between bg-transparent border border-gray-600" onClick={openModal}>
            <PiArrowsDownUp size={22} className="pt-1 mr-2" color="white"/>
            <span>{sortButtonTitle[sortCriteria]}</span>
          </button>
        </div>
      </div>

      {/* Libary Card List View  */}
      {activeTab === 'Authors' && bookAuthors?.allAuthors.length > 0 ? (
        <CardList allAuthors={bookAuthors.allAuthors} authorBooks={bookAuthors} />
      ) : activeTab === 'Genres' && bookGenres?.allGenres.length > 0 ? (
        <CardList allGenres={allGenres} genreBooks={genreBooks} />
      ) : (
        sortedBooks && sortedBooks.length > 0 && <CardList books={sortedBooks} isSearchPage={false} />
      )}

      {/* Snackbar */}
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
