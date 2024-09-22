import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Book } from '../../types/api';

interface CardProps {
  book: Book;
}

const BookshelfCard = ({ book }: CardProps) => {
  const { authors, imageLink, title } = book
  const navigate = useNavigate();
  const titleSubdomain = encodeURIComponent(title);

  const handleCardClick = () => {
    navigate(`/library/books/${titleSubdomain}`, { state: { book } });
  };

  return (
    <div className="card_wrapper bg-white relative box-border min-w-0 h-full cursor-pointer rounded-lg truncate dark:bg-maastricht" onClick={handleCardClick}>
      <div className="card_thumbnail rounded-md relative ">
        <div className="card_thumbnail_container relative rounded-md w-full">
          <img
            alt={`Book cover thumbnail for ${title}`}
            className="thumbnail absolute block top-0 left-0 h-full w-full rounded-md object-cover object-center cursor-pointer"
            loading="lazy"
            src={imageLink}
          />
        </div>
      </div>
      <div className="card_copy flex flex-col gap-x-4 items-baseline truncate mt-2">
        <div className="copy_container flex flex-col text-left">
          <a href="/" className="title text-charcoal w-full truncate dark:text-white">{title}</a>
          <div className="author_container">
            <span className="author_details">
              <div className="author_text text-sm text-gray-400 relative w-full">
                <a className="relative w-full text-cadet-gray" href="/">{authors[0]}</a>
              </div>
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}

export default React.memo(BookshelfCard);
