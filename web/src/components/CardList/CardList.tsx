import CardListItem from "./CardListItem"
import { Book } from '../../pages/library';

interface CardListItemProps {
  books: Book[];
}

export default function CardList({ books }: CardListItemProps) {

  return (
    <div className="card_list__wrapper flex flex-col relative w-full max-w-7xl mt-8">
      <ul className="flex flex-col justify-center rounded text-white">
          {books.map((book: Book) => (
            <CardListItem key={`${book.id}-${book.title}-${book.pageCount}`} book={book} />
          ))}
      </ul>
    </div>
  );
}
