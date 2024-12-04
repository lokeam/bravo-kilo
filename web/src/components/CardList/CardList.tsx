import CardListItem from './CardListItem'
import CardListItemAuthor from './CardListItemAuthor';
import CardListItemGenre from './CardListItemGenre';
import CardListItemTag from './CardListItemTag';
import { motion } from 'framer-motion';
import { Book, BookAuthorsData } from '../../types/api';

type CardListItemDefault = {
  books: Book[];
  isSearchPage?: boolean;
}

interface CardListItemAuthor {
  allAuthors: string[];
  authorBooks: BookAuthorsData;
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

interface CardListItemTag {
  allTags: string[];
  tagBooks: {
    [index: string]: {
      bookList: Book[];
      tagImgs: string[];
    };
  };
}

type CardListItemProps = CardListItemDefault | CardListItemAuthor | CardListItemGenre | CardListItemTag;

// Type Guard Checks
function isCardListItemDefault(props: CardListItemProps): props is CardListItemDefault {
  return Array.isArray((props as CardListItemDefault).books);
}

function isCardListItemAuthor(props: CardListItemProps): props is CardListItemAuthor {
  return 'allAuthors' in props && 'authorBooks' in props && Array.isArray((props as CardListItemAuthor).allAuthors);
}

function isCardListItemGenre(props: CardListItemProps): props is CardListItemGenre {
  return 'allGenres' in props && 'genreBooks' in props;
}

function isCardListItemTag(props: CardListItemProps): props is CardListItemTag {
  return 'allTags' in props && 'tagBooks' in props;
}

export default function CardList(props: CardListItemProps) {

  // Author Card List
  if (isCardListItemAuthor(props)) {
    console.log('CardList: Author section', {
      allAuthors: props.allAuthors,
      authorBooks: props.authorBooks
    });

    if (!props.allAuthors.length) {
      console.warn('No authors available');
      return null;
    }

    return (
      <motion.div
        className="card_list__wrapper pb-20 md:pb-4 flex flex-col relative w-full max-w-7xl mt-8"
        animate={{ opacity: 1 }}
        initial={{ opacity: 0 }}
        exit={{ opacity: 0 }}
        layout
      >
      <ul className="flex flex-col justify-center rounded text-white">
        {props.allAuthors.map((authorName: string, index: number) => {
          const books = props.authorBooks.byAuthor[authorName];
          console.log(`Processing author: ${authorName}`, {
            authorName,
            books,
            hasBooks: Boolean(books?.length)
          });

          // Type guard to ensure we're passing Book[] to CardListItemAuthor
          if (!Array.isArray(books) || !books.length || typeof books[0] === 'string') {
            return null;
          }
          return (
            <CardListItemAuthor
              key={`${authorName}-${index}`}
              authorName={authorName}
              books={books as Book[]}
            />
          );
        })}
      </ul>
      </motion.div>
    );
  }

  // Genre Card List
  if (isCardListItemGenre(props)) {
    console.log('Card List GENRE flag tripped');
    return (
      <motion.div
        className="card_list__wrapper pb-20 md:pb-4 flex flex-col relative w-full max-w-7xl mt-8"
        animate={{ opacity: 1 }}
        initial={{ opacity: 0 }}
        exit={{ opacity: 0 }}
        layout
      >
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
      </motion.div>
    );
  }


  // Tag Card List
  if (isCardListItemTag(props)) {
    return (
      <motion.div
        className="card_list__wrapper pb-20 md:pb-4 flex flex-col relative w-full max-w-7xl mt-8"
        animate={{ opacity: 1 }}
        initial={{ opacity: 0 }}
        exit={{ opacity: 0 }}
        layout
      >
        <ul className="flex flex-col justify-center rounded text-white">
          {props.allTags.map((tagName: string, index: number) => (
              <CardListItemTag
                key={tagName}
                tagImgs={props.tagBooks[String(index)].tagImgs || []}
                tagName={tagName}
                books={props.tagBooks[String(index)].bookList || []}
              />
          ))}
        </ul>
      </motion.div>
    );
  }

  // Standard Library.tsx Card List
  if (isCardListItemDefault(props)) {
    return (
      <motion.div
        className="card_list__wrapper pb-20 md:pb-4 flex flex-col relative w-full max-w-7xl mt-8"
        animate={{ opacity: 1 }}
        initial={{ opacity: 0 }}
        exit={{ opacity: 0 }}
        layout
      >
        <ul className="flex flex-col justify-center rounded text-white">
          {props.books.map((book: Book) => (
            <CardListItem key={`${book.id}-${book.title}-${book.pageCount}`} book={book} isSearchPage={props.isSearchPage} />
          ))}
        </ul>
      </motion.div>
    );
  }

  // Fallback if no condition is met
  return null;
}
