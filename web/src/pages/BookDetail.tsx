import { useLocation, useParams, useNavigate } from 'react-router-dom';
import ImagePlaceholder from '../components/CardList/ImagePlaceholder';

import useScrollShrink from "../hooks/useScrollShrink";
import useFetchBookById from '../hooks/useFetchBookById';
import { Book } from './Library';

import { IoIosAdd } from "react-icons/io";
import { TbEdit } from "react-icons/tb";


const BookDetail = () => {
  const { bookID } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const imageRef = useScrollShrink();
  const enabled = !location.state?.book;

  const { data: bookData, isLoading, isError } = useFetchBookById(bookID as string, enabled);
  const book: Book = bookData || location.state?.book;

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (isError) {
    return <div>Error fetching book details</div>;
  }

  if (!book) {
    return <div>No book found</div>;
  }

  console.log('book detail page, book data: ', book);

  const authors = book.authors?.join(', ') || 'Unknown Author';
  const genres = book.genres?.join(', ') || ['Unknown Genre'];

  return (
    <div className="bk_book_page mx-auto flex flex-col align-center max-w-screen-md h-screen p-5 pt-24">

      <div className="bk_book_thumb relative flex justify-center align-center rounded w-full">
        {
          book.imageLinks.length > 0 ? (
            <img
            alt={`Thumbnail for ${book.title}`}
            className="bk_book_thumb_img h-52 w-52"
            loading="lazy"
            src={book.imageLinks[0]}
            ref={imageRef}
          />
          ) : (
            <ImagePlaceholder isBookDetail />
          )
        }
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
        {
          !book.isInLibrary ? (
            <button onClick={() => navigate(`/library/books/${bookID}/edit`, { state: { book } })} className="flex items-center justify-center rounded bg-hepatica font-bold">
              <IoIosAdd className="h-8 w-8 mr-4"/>Add Book to Library
            </button>
          ) : (
            <button onClick={() => navigate(`/library/books/${bookID}/edit`, { state: { book } })} className="flex items-center justify-center rounded bg-hepatica font-bold">
              <TbEdit className="h-8 w-8 mr-4"/>Edit Book Details
            </button>
          )
        }
      </div>

      <div className="bk_book__details text-left my-4">
        <h3 className="text-2xl font-bold pb-2">Product Details</h3>
        <div className="bk_book_metadata flex flex-col mb-4">
          <p className="my-1 text-cadet-gray"><span className="font-bold mr-1">Publish Date:</span>{book.publishDate !== "" ? book.publishDate : "No publish date available"}</p>
          <p className="my-1 text-cadet-gray"><span className="font-bold mr-1">Pages:</span>{book.pageCount !== 0 ? book.pageCount : "No page count available"}</p>
          <p className="my-1 text-cadet-gray"><span className="font-bold mr-1">Language:</span>{book.language !== "" ? book.language : "No language classification available"}</p>
          <p className="my-1 text-cadet-gray"><span className="font-bold mr-1">ISBN-10:</span>{book.isbn10 !== "" ? book.isbn10 : "No ISBN10 data available"}</p>
          <p className="my-1 text-cadet-gray"><span className="font-bold mr-1">ISBN-13:</span>{book.isbn13 !== "" ? book.isbn13: "No ISBN13 data available"}</p>
        </div>
        <div className="bk_book__details flex flex-col text-left mb-4">
          <h3 className="text-2xl font-bold mb-4">Tagged as:</h3>
          <div className="bk_book_genres w-full flex flex-row items-center content-evenly gap-6">
            {book.genres.map((genre) => (
              <button key={`${genre}}`} className="border border-gray-500">{genre}</button>
            ))}
          </div>
        </div>
        {
          book.notes !== "" ? (
            <div className="bk_description text-left mb-4">
              <h3 className="text-2xl font-bold mb-2">Personal Notes</h3>
              <p className="text-cadet-gray">{book.notes}</p>
            </div>
          ) : null
        }
        <div className="bk_description text-left mb-4">
          <h3 className="text-2xl font-bold pb-2">Book Description</h3>
          <p className="text-cadet-gray">{book.description !== "" ? book.description : "No book description available"}</p>
        </div>
      </div>
    </div>
  );
};

export default BookDetail;
