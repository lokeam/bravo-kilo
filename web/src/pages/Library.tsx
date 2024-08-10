import { useCallback, useEffect, useMemo, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import useStore from '../store/useStore';
import useFetchBooks from '../hooks/useFetchBooks';
import CardList from '../components/CardList/CardList';
import Modal from '../components/Modal/Modal';
import '../components/Modal/Modal.css';
import { PiArrowsDownUp } from 'react-icons/pi';
import { fetchBooksAuthors, fetchBooksFormat, fetchBooksGenres } from '../service/apiClient.service';

export interface Book {
  id?: number;
  title: string;
  subtitle?: string;
  description?: string;
  language: string;
  pageCount: number;
  publishDate?: string;
  authors: string[];
  imageLinks: string[];
  genres: string[];
  notes?: string;
  formats?: ('physical' | 'eBook' | 'audioBook')[];
  createdAt?: string;
  lastUpdated?: string;
  isbn10: string;
  isbn13: string;
  isInLibrary?: boolean;
}

// Create Book Authors/Genres intersection types that combines both obj types, expect both types
type BookAuthorsData = {
  allAuthors: string[]
} & {
  [index: string]: Book[];
};

type BookGenresData = {
  allGenres: string[]
} & {
  [index: string]: {
    bookList: Book[];
    genreImgs: string[];
  };
};

const Library = () => {
  const [opened, setOpened] = useState(false);
  const { search } = useLocation();
  const query = new URLSearchParams(search);
  const userID = parseInt(query.get('userID') || '0', 10);
  const { sortCriteria, sortOrder, setSort, activeTab, setActiveTab } = useStore();
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
  const {
    data: bookAuthors,
    isLoading: isAuthorsLoading,
    isError: isAuthorsError,
  } = useQuery<BookAuthorsData>({
    queryKey: ['userBookAuthors', userID],
    queryFn: () => fetchBooksAuthors(userID),
    staleTime: Infinity,
    gcTime: 1000 * 60 * 60 * 24,
  });

  // Use useQuery to get cached book genres
  const {
    data: bookGenres,
    isLoading: isGenresLoading,
    isError: isGenresError,
  } = useQuery<BookGenresData>({
    queryKey: ['userBookGenres', userID],
    queryFn: () => fetchBooksGenres(userID),
    staleTime: Infinity,
    gcTime: 1000 * 60 * 60 * 24,
  });

  useEffect(() => {
    if (bookAuthors) {
      console.log('Authors data:', bookAuthors); // Log to verify authors data
    }
  }, [bookAuthors]);

  useEffect(() => {
    if (bookGenres) {
      console.log('Genres data:', bookGenres); // Log to verify genres data
    }
  }, [bookGenres]);

  // Retrieve cached books formats
  const bookFormats = queryClient.getQueryData<{
    audioBooks: Book[],
    eBooks: Book[],
    physicalBooks: Book[]
  }>(['booksFormat', userID]);


  console.log('bookAuthors: ', bookAuthors);
  console.log('****************************************');
  console.log('bookGenres: ', bookGenres);

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

          case "publishDate":
            return order === "asc"
              ? new Date(a.publishDate).getTime() - new Date(b.publishDate).getTime()
              : new Date(b.publishDate).getTime() - new Date(a.publishDate).getTime();

          case "author": {
            const aSurname = a.authors[0].split(" ").pop() || "";
            const bSurname = b.authors[0].split(" ").pop() || "";
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
    let booksToSort = [];
    if (activeTab === 'Audiobooks' && bookFormats) {
      console.log('Using audiobooks format');
      booksToSort = bookFormats.audioBooks;
    } else if (activeTab === 'eBooks' && bookFormats) {
      console.log('Using eBooks format');
      booksToSort = bookFormats.eBooks;
    } else if (activeTab === 'Printed Books' && bookFormats) {
      console.log('Using printed books format');
      booksToSort = bookFormats.physicalBooks;
    } else  {
      // Default to all books
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

      // Navigate to Author page when Authors tab clicked
      // if (tab === 'Authors') {
      //   navigate('/library/author');
      // }
    },
    [setActiveTab]
  );

  if (isLoading || isAuthorsLoading || isGenresLoading) {
    return <div>Loading...</div>;
  }

  if (isError || isAuthorsError || isGenresError) {
    return <div>Error loading books</div>;
  }

  const sortButtonTitle = {
    'title': 'Title: A to Z',
    'author': 'Author: A to Z',
    'publishDate': 'Release date: New to Old',
    'pageCount': 'Page count: Short to Long',
  };

  const openModal = () => setOpened(true);
  const closeModal = () => setOpened(false);

  return (
    <div className="bk_lib flex flex-col items-center place-content-around px-5 antialiased md:px-1 md:ml-24 h-screen pt-28">
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
      {activeTab === 'Authors' && bookAuthors?.allAuthors ? (
        <CardList allAuthors={bookAuthors.allAuthors} authorBooks={bookAuthors} />
      ) : activeTab === 'Genres' && bookGenres?.allGenres ? (
        <CardList allGenres={bookGenres.allGenres} genreBooks={bookGenres} />
      ) : (
        sortedBooks && sortedBooks.length > 0 && <CardList books={sortedBooks} isSearchPage={false} />
      )}
    </div>
  )
}

export default Library;
