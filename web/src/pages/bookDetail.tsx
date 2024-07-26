import { useLocation, useParams, useNavigate } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';

import { IoArrowBackCircle } from "react-icons/io5";
import { BsThreeDotsVertical } from "react-icons/bs";
import { TbEdit } from "react-icons/tb";
import useScrollShrink from "../hooks/useScrollShrink";

import axios from 'axios';

interface Book {
  id: number;
  title: string;
  subtitle?: string;
  description: string;
  language: string;
  pageCount: number;
  publishDate: string;
  authors: string[];
  imageLinks: string[];
  genres: string[];
  notes: string;
  formats: ('physical' | 'eBook' | 'audioBook')[];
  createdAt: string;
  lastUpdated: string;
  isbn10: string;
  isbn13: string;
}

const fetchBook = async (bookID: string) => {
  const { data } = await axios.get(`${import.meta.env.VITE_API_ENDPOINT}/api/v1/books/${bookID}`, {
    withCredentials: true
  });
  return data.book;
};

const BookDetail = () => {
  const { bookID } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const imageRef = useScrollShrink();

  const [book, setBook] = useState<Book | null>(null);

  const { data, isLoading, isError } = useQuery({
    queryKey: ['book', bookID],
    queryFn: () => fetchBook(bookID as string),
    enabled: !location.state?.book,
  });

  useEffect(() => {
    if (location.state?.book) {
      setBook(location.state.book);
    } else if (data) {
      setBook(data);
    }
  }, [data, location.state?.book]);

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

  return (
    <div className="bk_book_page mx-auto flex flex-col align-center max-w-screen-md h-screen p-5">
      <div className="bk_book_page__top_nav flex flex-row sticky top-0 justify-between w-full">
        <button onClick={() => navigate('/library')} className="bg-transparent h-auto w-auto">
          <IoArrowBackCircle color="white" className="h-12 w-12" />
        </button>
        <button className="bg-transparent">
          <BsThreeDotsVertical className="h-12 w-12" />
        </button>
      </div>

      <div className="bk_book_thumb relative flex justify-center align-center rounded w-full">
        <img
          alt={`Thumbnail for ${book.title}`}
          className="bk_book_thumb_img h-52 w-52"
          loading="lazy"
          src={book.imageLinks[0]}
          ref={imageRef}
        />
      </div>

      <div className="bk_book_title_wrapper flex flex-col justify-center relative mt-6 mb-2 text-center">
        <h1 className="text-5xl text-center mb-2">{book.title}</h1>
        {book.subtitle && <h2 className="text-2xl">{book.subtitle}</h2>}
      </div>

      <div className="bk_book_metadata my-3">
        <div className="text-sm font-bold">BY</div>
        <div className="">{book.authors.join(', ')}</div>
        <div className="">{book.genres.join(', ')}</div>
      </div>

      <div className="bk_book_cta flex flex-col w-full my-3">
        <button className="flex items-center justify-center rounded bg-hepatica font-bold">
          <TbEdit className="h-8 w-8 mr-4"/>Edit Information
        </button>
      </div>

      <div className="bk_book__details text-left my-4">
        <h3 className="text-2xl font-bold pb-2">Product Details</h3>
        <div className="bk_book_metadata flex flex-col mb-4">
          <p className="my-1 text-cadet-gray"><span className="font-bold mr-1">Publish Date:</span>{book.publishDate}</p>
          <p className="my-1 text-cadet-gray"><span className="font-bold mr-1">Pages:</span>{book.pageCount}</p>
          <p className="my-1 text-cadet-gray"><span className="font-bold mr-1">Language:</span>{book.language}</p>
          <p className="my-1 text-cadet-gray"><span className="font-bold mr-1">ISBN-10:</span>{book.isbn10}</p>
          <p className="my-1 text-cadet-gray"><span className="font-bold mr-1">ISBN-13:</span>{book.isbn13}</p>
        </div>
        <div className="bk_book__details flex flex-col text-left mb-4">
          <h3 className="text-2xl font-bold mb-4">Tagged as:</h3>
          <div className="bk_book_genres w-full flex flex-row items-center content-evenly gap-6">
            {book.genres.map((genre) => (
              <button key={`${genre}}`} className="border border-gray-500">{genre}</button>
            ))}
          </div>
        </div>
        <div className="bk_description text-left mb-4">
          <h3 className="text-2xl font-bold mb-2">Personal Notes</h3>
          <p className="text-cadet-gray">TBI - Etiam eget erat pulvinar, aliquam enim a, feugiat lorem. Nulla congue purus quis suscipit suscipit. In auctor magna in tempor lacinia. Vivamus leo ligula, commodo sit amet porttitor id, placerat ac justo. Curabitur non ultrices tellus, a porttitor felis. Maecenas quis interdum urna. Pellentesque sed urna quis tortor feugiat semper. Etiam laoreet ut ex in pellentesque. Phasellus eu lectus porttitor, scelerisque justo vitae, euismod ligula.</p>
        </div>
        <div className="bk_description text-left mb-4">
          <h3 className="text-2xl font-bold pb-2">Book Description</h3>
          <p className="text-cadet-gray">TBI - Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem. Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit laboriosam, nisi ut aliquid ex ea commodi consequatur? Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur, vel illum qui dolorem eum fugiat quo voluptas nulla pariatur.</p>
        </div>
      </div>
    </div>
  );
};

export default BookDetail;