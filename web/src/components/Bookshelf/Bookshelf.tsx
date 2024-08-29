import BookshelfHeader from './BookshelfHeader';
import BookshelfCardList from './BookshelfCardList';
import { Book } from '../../types/api';

import './Bookshelf.css';

interface BookShelfProps {
  category?: string;
  isLoading: boolean;
  books: Book[];
}

const Bookshelf = ({ books, category, isLoading }: BookShelfProps) => {
  const lastUpdatedBooks = books?.slice().sort((a, b) => {
    return new Date(b.lastUpdated).getTime() - new Date(a.lastUpdated).getTime();
  });

  return (
    <section className="bookshelf_wrapper pb-4 md:pb-4 flex flex-col relative w-full max-w-7xl">
      <BookshelfHeader heading={category} />
      <BookshelfCardList cardData={lastUpdatedBooks} isLoading={isLoading} />
    </section>
  )
};

export default Bookshelf;
