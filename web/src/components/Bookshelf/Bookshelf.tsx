import BookshelfHeader from './BookshelfHeader';
import BookshelfCardList from './BookshelfCardList';
import { Book } from '../../pages/Library';
import './Bookshelf.css';

interface BookShelfProps {
  category?: string;
  isLoading: boolean;
  books: Book[];
}

const Bookshelf = ({ books, category, isLoading }: BookShelfProps) => {
  console.log('bookshelf render');
  const lastUpdatedBooks = books?.slice().sort((a, b) => {
    return new Date(b.publishDate).getTime() - new Date(a.publishDate).getTime();
  });

  console.log('Bookshelf category: ', category);
  console.log('Bookshelf books: ', books);

  return (
    <section className="bookshelf_wrapper pb-20 md:pb-4 flex flex-col relative w-full max-w-7xl">
      <BookshelfHeader heading={category} showAllUrl="/" />
      <BookshelfCardList cardData={lastUpdatedBooks} isLoading={isLoading} />
    </section>
  )
};

export default Bookshelf;
