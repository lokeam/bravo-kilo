import CardListItem from "./CardListItem"
import { PiArrowsDownUp } from "react-icons/pi";
import { Book } from '../../pages/library';

interface CardListItemProps {
  books: Book[];
}

export default function CardList({ books }: CardListItemProps) {

  return (
    <div className="card_list__wrapper max-w-5xl min-w-[800px] mt-8">
      <div className="flex flex-row justify-between items-center text-left text-white border-b-2 border-solid border-zinc-700 pb-6 mb-2">
        <div className="mt-1">{books.length} volumes</div>
        <div className="flex flex-row">
          <button className="flex flex-row bg-transparent mr-1">
            <PiArrowsDownUp className="w-5 h-5 pt-1 mr-2" color="white"/>
              Recent activity
            </button>
          <button className="bg-transparent">Select</button>
        </div>
      </div>
      <ul className="flex flex-col justify-center rounded text-white">
          {books.map((book: Book) => (
            <CardListItem key={`${book.id}-${book.title}-${book.pageCount}`} book={book} />
          ))}
      </ul>
    </div>
  );
}
