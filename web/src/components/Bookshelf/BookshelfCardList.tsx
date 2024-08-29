import React from 'react';
import BookshelfCard from './BookshelfCard';
import { Book } from '../../types/api';

interface BookshelfListProps {
  cardData: Book[];
  isLoading: boolean;
}

const BookshelfCardList = ({ cardData = [], isLoading }: BookshelfListProps) => {
  console.log('cardData: ', cardData);
  console.log('isLoading: ', isLoading);

  return (
    <div className="bookshelf_body relative w-full z-10 pb-8">
      <div className="bookshelf_grid_wrapper box-border ">
        <div className="bookshelf_grid_body box-content overflow-visible w-full">
          <ul className="bookshelf_grid box-border grid grid-flow-col items-stretch gap-x-2.5 overflow-x-auto overflow-y-auto overscroll-x-none scroll-smooth snap-start snap-x snap-mandatory list-none m-0 p-0">
            {
              isLoading ? (
                <div>Loading data...</div>
              ) : cardData.length > 0 ? (
                cardData.map((book) => (
                  <li key={`${book.id}-${book.title}`}>
                    <BookshelfCard
                      book={book}
                    />
                  </li>
                ))
              ) : (
                <li>No books found</li>
              )
            }
          </ul>
        </div>
      </div>
    </div>
  );
};

export default React.memo(BookshelfCardList);
