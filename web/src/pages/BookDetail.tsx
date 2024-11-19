import { useLocation, useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import ImagePlaceholder from '../components/CardList/ImagePlaceholder';
import { fetchBookIDByTitle } from '../service/apiClient.service';
import useScrollShrink from "../hooks/useScrollShrink";
import PageWithErrorBoundary from '../components/ErrorMessages/PageWithErrorBoundary';
import { useThemeStore } from '../store/useThemeStore';
import { displayPublishDate } from '../utils/displayPublishDate';
import Loading from '../components/Loading/Loading';
import { IoIosAdd } from "react-icons/io";
import { IoIosWarning } from "react-icons/io";
import { TbEdit } from "react-icons/tb";
import QuillContent from '../components/Quill/QuillContent';

interface MissingInfoWarningProps {
  emptyFields: string[]
}

const MissingInfoWarning = ({emptyFields}: MissingInfoWarningProps) => {
  return (
    <div className="p-4 mb-4 border border-orange-500 text-orange-500 dark:border-yellow-300 dark:text-yellow-300 rounded-lg ">
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
        <p className="text-orange-500 dark:text-yellow-300">You may enter placeholder data if you want to save this book in your library, but you'll need the official information if you want your entry to be identifiable by search.</p>
      </div>
    </div>
  );
}

const BookDetail = () => {
  const { bookTitle } = useParams();
  const decodedTitle = bookTitle ? decodeURIComponent(bookTitle.toLowerCase()) : '';
  const imageRef = useScrollShrink();
  const navigate = useNavigate();
  const location = useLocation();
  const { theme } = useThemeStore();

  const bookFromState = location.state?.book;
  const isInLibrary = bookFromState?.isInLibrary ?? true;
  const isDarkMode = theme === 'dark';

  console.log('Rendering BookDetail component');
  console.log('Decoded title:', decodedTitle);
  console.log('bookFromState:', bookFromState);
  console.log('isInLibrary:', isInLibrary);
  console.log('enabled:', !!decodedTitle && !isInLibrary);
  console.log('isDarkMode:', isDarkMode);


  const { data: bookID, isLoading, isError } = useQuery({
    queryKey: ['bookID', decodedTitle],
    queryFn: () => {
      console.log('Fetching book ID with title:', decodedTitle);
      return fetchBookIDByTitle(decodedTitle);
    },
    enabled: isInLibrary,
  });

  console.log('Fetched bookID:', bookID);
  console.log('Loading status:', isLoading);
  console.log('Error status:', isError);

  // Display loading state
  if (isLoading) {
    return <Loading />;
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

  const bookCover = book?.imageLink;
  console.log('testing bookCover: ', bookCover);

  const parseRichTextContent = (content: any) => {
    if (!content) return null;

    try {
      // If content is already a Delta-like obj
      if (
        content &&
        typeof content === 'object' &&
        Array.isArray(content.ops)
      ) {
        return content;
      }

      // If content is a stringified Delta obj
      if (typeof content === 'string') {
        const parsedContent = JSON.parse(content);

        if (parsedContent && Array.isArray(parsedContent.ops)) {
          return parsedContent;
        }
      }

      return null;
    } catch (error) {
      console.error('parseRichTextContent error: ', error);
      return null;
    }
  }

  const description = parseRichTextContent(book.description);
  const notes = parseRichTextContent(book.notes);

  console.log('book.description: ', description);
  console.log('notes: ', notes);

  const hasNotes = notes.ops && notes.ops.length > 0;

  return (
    <PageWithErrorBoundary fallbackMessage="Error loading book detail page">
      <div className="bk_edit_book_page_wrapper bg-cover min-h-screen overflow-x-hidden overflow-y-auto z-10 bg-white-smoke dark:bg-transparent relative flex flex-col items-center place-content-around px-5 antialiased mdTablet:pr-5 mdTablet:ml-24">
        <div className={`book_coverImage ${!isDarkMode ? '' : 'blurBg'}  -z-10`}></div>

        <div className="bk_edit_book_page max-w-screen-mdTablet pb-20 md:pb-4 flex flex-col relative w-full">
          <div className="bk_book_thumb relative flex justify-center align-center rounded w-full">
            {book.imageLink && book.imageLink !== "" ? (
              <img
                alt={`Thumbnail for ${book.title}`}
                className="bk_book_thumb_img object-contain rounded w-52"
                loading="lazy"
                src={book.imageLink}
                ref={imageRef}
                onLoad={() => console.log('Image loaded, ref: ', imageRef.current)}
              />
            ) : (
              <ImagePlaceholder isBookDetail />
            )}
          </div>

          <div className="bk_book_title_wrapper flex flex-col justify-center relative mt-6 mb-2 text-center">
            <h1 className="text-5xl text-center mb-2 text-black dark:text-white">{book.title}</h1>
            {book.subtitle && <h2 className="text-2xl text-charcoal dark:text-az-white">{book.subtitle}</h2>}
          </div>

          <div className="bk_book_metadata my-3">
            <div className="text-sm font-bold text-charcoal dark:text-az-white">BY</div>
            <div className="text-charcoal font-extrabold dark:text-az-white">{authors}</div>
            <div className=" text-charcoal dark:text-az-white">{genres}</div>
          </div>

          <div className="bk_book_cta flex flex-col w-full my-3">
            {!book.isInLibrary ? (
              <button
                onClick={() =>
                  navigate(`/library/books/add/search`, {
                    state: {
                      book: {
                        ...book,
                        description: JSON.stringify(book?.description),
                        notes: JSON.stringify(book?.notes),
                      }
                    }
                  })
                }
                className="flex items-center justify-center rounded border font-bold bg-vivid-blue hover:bg-vivid-blue-d dark:border-vivid-blue dark:bg-vivid-blue dark:hover:bg-vivid-blue-d dark:hover:border-vivid-blue-d transition duration-500 ease-in-out"
              >
                <IoIosAdd className="h-8 w-8 mr-4" />
                Add Book to Library
              </button>
            ) : (
              <button
                onClick={() =>
                  navigate(`/library/books/${bookID}/edit`, {
                    state: {
                      book: {
                        ...book,
                        description: JSON.stringify(book?.description),
                        notes: JSON.stringify(book?.notes),
                      }
                    }
                  })
                }
                className="flex items-center justify-center rounded font-bold bg-vivid-blue hover:bg-vivid-blue-d dark:via-vivid-blue dark:hover:via-vivid-blue-d dark:hover:border-vivid-blue-d transition duration-500 ease-in-out"
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
            <h3 className="text-2xl font-bold pb-2 text-black dark:text-white">Product Details</h3>
            <div className="bk_book_metadata flex flex-col mb-4">
              <p className="my-1 text-charcoal dark:text-cadet-gray">
                <span className="font-bold mr-1">Publish Date:</span>
                {book.publishDate !== ''
                  ? displayPublishDate(book.publishDate)
                  : 'No publish date available'}
              </p>
              <p className="my-1 text-charcoal dark:text-cadet-gray">
                <span className="font-bold mr-1">Pages:</span>
                {book.pageCount !== 0 ? book.pageCount : 'No page count available'}
              </p>
              <p className="my-1 text-charcoal dark:text-cadet-gray">
                <span className="font-bold mr-1">Language:</span>
                {book.language !== ''
                  ? book.language
                  : 'No language classification available'}
              </p>
              <p className="my-1 text-charcoal dark:text-cadet-gray">
                <span className="font-bold mr-1">ISBN-10:</span>
                {book.isbn10 !== '' ? book.isbn10 : 'No ISBN10 data available'}
              </p>
              <p className="my-1 text-charcoal dark:text-cadet-gray">
                <span className="font-bold mr-1">ISBN-13:</span>
                {book.isbn13 !== '' ? book.isbn13 : 'No ISBN13 data available'}
              </p>
            </div>
            <div className="bk_book__details flex flex-col text-left mb-4">
              <h3 className="text-2xl font-bold mb-4 text-black dark:text-white">Genres:</h3>
              <div className="bk_book_genres w-full flex flex-row flex-wrap items-center content-evenly gap-6">
                {book.genres && book.genres.length > 0 && book.genres.map((genre: string, index: number) => (
                  <button key={`${genre}-${index}`} className="bg-white-smoke dark:bg-black text-black dark:text-white border border-gray-500 cursor-default hover:border-strong-violet dark:hover:border-strong-violet transition duration-500 ease-in-out">
                    {genre}
                  </button>
                ))}
              </div>
            </div>
            <div className="bk_book__details flex flex-col text-left mb-4">
              <h3 className="text-2xl font-bold mb-4 text-black dark:text-white">Assigned Personal Tags:</h3>
              <div className="bk_book_genres w-full flex flex-row flex-wrap items-center content-evenly gap-6">
                {book.tags && book.tags.length > 0 && book.tags.map((tag: string, index: number) => (
                  <button key={`${tag}-${index}`} className="bg-white-smoke dark:bg-black text-black dark:text-white border border-gray-500 cursor-default hover:border-strong-violet dark:hover:border-strong-violet transition duration-500 ease-in-out">
                    {tag}
                  </button>
                ))}
              </div>
            </div>
            <div className="bk_description text-left mb-4">
              <h3 className="text-2xl font-bold mb-4 pb-2 text-black dark:text-white">Book Description</h3>
              { description && <QuillContent content={description} /> }
            </div>
            {hasNotes && (
              <div className="bk_description text-left mb-4">
                <h3 className="text-2xl font-bold mb-4 text-black dark:text-white">Personal Notes</h3>
                { notes && <QuillContent content={notes} />}
              </div>
            )}
          </div>
        </div>
      </div>
    </PageWithErrorBoundary>

  );
};

export default BookDetail;

