import { useLocation, useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import ImagePlaceholder from '../components/CardList/ImagePlaceholder';
import { fetchBookIDByTitle } from '../service/apiClient.service';
import useScrollShrink from "../hooks/useScrollShrink";
import { IoIosAdd } from "react-icons/io";
import { IoIosWarning } from "react-icons/io";
import { TbEdit } from "react-icons/tb";

interface MissingInfoWarningProps {
  emptyFields: string[]
}

const MissingInfoWarning = ({emptyFields}: MissingInfoWarningProps) => {
  return (
    <div className="p-4 mb-4 border border-yellow-300 rounded-lg text-yellow-300">
      <div className="bk_book_metadata flex flex-col">
        <div className="flex flex-row pb-3">
          <IoIosWarning size={25} className='mr-2'/>
          <span>This book listing missing the following data:</span>
        </div>
        <ul className="list-disc pl-5 pb-3">
          {emptyFields.map((emptyFieldStr, index) => (
            <li key={`${emptyFieldStr}-${index}`} className="font-bold">{emptyFieldStr}</li>
          ))}
        </ul>
        <p className="text-yellow-300">You may enter placeholder data if you want to save this book in your library, but you'll need the official information if you want your entry to be identifiable by search.</p>
      </div>
    </div>
  );
}

const BookDetail = () => {
  console.log('Rendering BookDetail component');
  const { bookTitle } = useParams(); // Access bookTitle from URL
  const decodedTitle = bookTitle ? decodeURIComponent(bookTitle) : '';
  const imageRef = useScrollShrink();
  const navigate = useNavigate();
  const location = useLocation();



  const bookFromState = location.state?.book;
  const isInLibrary = bookFromState?.isInLibrary ?? true;

  console.log('Rendering BookDetail component');
  console.log('Decoded title:', decodedTitle);
  console.log('bookFromState:', bookFromState);
  console.log('isInLibrary:', isInLibrary);
  console.log('enabled:', !!decodedTitle && !isInLibrary);


  const { data: bookID, isLoading, isError } = useQuery({
    queryKey: ['bookID', decodedTitle],
    queryFn: () => {
      console.log('Fetching book ID with title:', decodedTitle);
      return fetchBookIDByTitle(decodedTitle);
    },
    enabled: isInLibrary,
    onSuccess: (data) => {
      console.log('Successfully fetched book ID:', data);
    },
    onError: (error) => {
      console.error('Error fetching book ID:', error);
    },
  });

  console.log('Fetched bookID:', bookID);
  console.log('Loading status:', isLoading);
  console.log('Error status:', isError);

  // Display loading state
  if (isLoading) {
    return <div>Loading...</div>;
  }

  // Display error message
  if (isError) {
    return <div>Error fetching book details</div>;
  }

  const book = bookFromState || {};
  console.log('Book data:', book);

  if (!book) {
    return <div>No book data available</div>;
  }

  const authors = book.authors?.join(', ') || 'Unknown Author';
  const genres = book.genres?.join(', ') || ['Unknown Genre'];

  return (
    <div className="bk_book_page mx-auto flex flex-col align-center max-w-screen-md h-screen p-5 pt-24">
      <div className="bk_book_thumb relative flex justify-center align-center rounded w-full">
        {book.imageLinks && book.imageLinks.length > 0 ? (
          <img
            alt={`Thumbnail for ${book.title}`}
            className="bk_book_thumb_img h-52 w-52"
            loading="lazy"
            src={book.imageLinks[0]}
            ref={imageRef}
          />
        ) : (
          <ImagePlaceholder isBookDetail />
        )}
      </div>

      <div className="bk_book_title_wrapper flex flex-col justify-center relative mt-6 mb-2 text-center">
        <h1 className="text-5xl text-center mb-2">{book.title}</h1>
        {book.subtitle && <h2 className="text-2xl">{book.subtitle}</h2>}
      </div>

      <div className="bk_book_metadata my-3">
        <div className="text-sm font-bold">BY</div>
        <div className="">{authors}</div>
        <div className="">{genres}</div>
      </div>

      <div className="bk_book_cta flex flex-col w-full my-3">
        {!book.isInLibrary ? (
          <button
            onClick={() =>
              navigate(`/library/books/add/search`, { state: { book } })
            }
            className="flex items-center justify-center rounded bg-hepatica font-bold"
          >
            <IoIosAdd className="h-8 w-8 mr-4" />
            Add Book to Library
          </button>
        ) : (
          <button
            onClick={() =>
              navigate(`/library/books/${bookID}/edit`, { state: { book } })
            }
            className="flex items-center justify-center rounded bg-hepatica font-bold"
          >
            <TbEdit className="h-8 w-8 mr-4" />
            Edit Book Information
          </button>
        )}
      </div>

      <div className="bk_book__details text-left my-4">
        {book.hasEmptyFields && (
          <MissingInfoWarning emptyFields={book.emptyFields} />
        )}
        <h3 className="text-2xl font-bold pb-2">Product Details</h3>
        <div className="bk_book_metadata flex flex-col mb-4">
          <p className="my-1 text-cadet-gray">
            <span className="font-bold mr-1">Publish Date:</span>
            {book.publishDate !== ''
              ? book.publishDate
              : 'No publish date available'}
          </p>
          <p className="my-1 text-cadet-gray">
            <span className="font-bold mr-1">Pages:</span>
            {book.pageCount !== 0 ? book.pageCount : 'No page count available'}
          </p>
          <p className="my-1 text-cadet-gray">
            <span className="font-bold mr-1">Language:</span>
            {book.language !== ''
              ? book.language
              : 'No language classification available'}
          </p>
          <p className="my-1 text-cadet-gray">
            <span className="font-bold mr-1">ISBN-10:</span>
            {book.isbn10 !== '' ? book.isbn10 : 'No ISBN10 data available'}
          </p>
          <p className="my-1 text-cadet-gray">
            <span className="font-bold mr-1">ISBN-13:</span>
            {book.isbn13 !== '' ? book.isbn13 : 'No ISBN13 data available'}
          </p>
        </div>
        <div className="bk_book__details flex flex-col text-left mb-4">
          <h3 className="text-2xl font-bold mb-4">Tagged as:</h3>
          <div className="bk_book_genres w-full flex flex-row items-center content-evenly gap-6">
            {book.genres.map((genre) => (
              <button key={`${genre}}`} className="border border-gray-500">
                {genre}
              </button>
            ))}
          </div>
        </div>
        {book.notes !== '' ? (
          <div className="bk_description text-left mb-4">
            <h3 className="text-2xl font-bold mb-2">Personal Notes</h3>
            <p className="text-cadet-gray">{book.notes}</p>
          </div>
        ) : null}
        <div className="bk_description text-left mb-4">
          <h3 className="text-2xl font-bold pb-2">Book Description</h3>
          <p className="text-cadet-gray">
            {book.description !== ''
              ? book.description
              : 'No book description available'}
          </p>
        </div>
      </div>
    </div>
  );
};

export default BookDetail;

