import CardListItem from './CardListItem'
import CardListItemAuthor from './CardListItemAuthor';
import CardListItemGenre from './CardListItemGenre';
import { Book } from '../../pages/Library';


interface CardListItemDefault {
  books: Book[];
  isSearchPage?: boolean;
}

interface CardListItemAuthor {
  allAuthors: string[];
  authorBooks: {
    [index: string]: Book[]
  };
}

interface CardListItemGenre {
  allGenres: string[];
  genreBooks: {
    [index: string]: {
      bookList: Book[];
      genreImgs: string[];
    };
  };
}

type CardListItemProps = CardListItemDefault | CardListItemAuthor | CardListItemGenre;

// Type Guard Checks
function isCardListItemDefault(props: CardListItemProps): props is CardListItemDefault {
  return Array.isArray((props as CardListItemDefault).books);
}

function isCardListItemAuthor(props: CardListItemProps): props is CardListItemAuthor {
  return 'allAuthors' in props && 'authorBooks' in props;
}

function isCardListItemGenre(props: CardListItemProps): props is CardListItemGenre {
  return 'allGenres' in props && 'genreBooks' in props;
}



export default function CardList(props: CardListItemProps) {
  // console.log('CardList Props: ', props);

  // Author Card List
  if (isCardListItemAuthor(props)) {
    //console.log('Card List Author flag tripped');
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

  // Genre Card List
  if (isCardListItemGenre(props)) {
    console.log('Card List GENRE flag tripped');
    return (
      <div className="card_list__wrapper pb-20 md:pb-4 flex flex-col relative w-full max-w-7xl mt-8">
        <ul className="flex flex-col justify-center rounded text-white">
          {props.allGenres.map((genreName: string, index: number) => (
              <CardListItemGenre
                key={genreName}
                genreImgs={props.genreBooks[String(index)].genreImgs || []}
                genreName={genreName}
                books={props.genreBooks[String(index)].bookList || []}
              />
          ))}
        </ul>
      </div>
    );
  }

  // Standard Library.tsx Card List
  if (isCardListItemDefault(props)) {
    //console.log('Card List Default flag tripped');
    return (
      <div className="card_list__wrapper pb-20 md:pb-4 flex flex-col relative w-full max-w-7xl mt-8">
        <ul className="flex flex-col justify-center rounded text-white">
          {props.books.map((book: Book) => (
            <CardListItem key={`${book.id}-${book.title}-${book.pageCount}`} book={book} isSearchPage={props.isSearchPage} />
          ))}
        </ul>
      </div>
    );
  }

  // Fallback if no condition is met
  return null;
}
