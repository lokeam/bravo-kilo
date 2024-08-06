import CardListItem from './CardListItem'
import CardListItemAuthor from './CardListItemAuthor';
import { Book } from '../../pages/Library';

interface CardListItemDefault {
  books: Book[];
}

interface CardListItemAuthor {
  allAuthors: string[];
  authorBooks: {
    [index: string]: Book[]
  };
}

type CardListItemProps = CardListItemDefault | CardListItemAuthor;

// Type Guard Checks
function isCardListItemDefault(props: CardListItemProps): props is CardListItemDefault {
  return Array.isArray((props as CardListItemDefault).books) && (props as CardListItemDefault).books.length > 0;
}

function isCardListItemAuthor(props: CardListItemAuthor): props is CardListItemAuthor {
  return 'allAuthors' in props;
}

export default function CardList(props: CardListItemProps) {

  console.log('CardList Props: ', props);

  // Standard Library.tsx Card List
  if (isCardListItemDefault(props)) {
    return (
      <div className="card_list__wrapper pb-20 md:pb-4 flex flex-col relative w-full max-w-7xl mt-8">
        <ul className="flex flex-col justify-center rounded text-white">
          {props.books.map((book: Book) => (
            <CardListItem key={`${book.id}-${book.title}-${book.pageCount}`} book={book} />
          ))}
        </ul>
      </div>
    );
  }

  // Author/Genre Card List
  if (isCardListItemAuthor(props)) {
    return (
      <div className="card_list__wrapper pb-20 md:pb-4 flex flex-col relative w-full max-w-7xl mt-8">
        <ul className="flex flex-col justify-center rounded text-white">
          {props.allAuthors.map((authorName: string, index: number) => (
            <CardListItemAuthor
              key={`${authorName}-${index}`}
              authorName={authorName}
              books={props.authorBooks[String(index)] || []}
            />
          ))}
        </ul>
      </div>
    );
  }

  // Fallback if neither condition is met
  return null;
}
