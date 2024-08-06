import { useParams, useLocation } from "react-router-dom";
import { Book } from "./Library";
import CardList from "../components/CardList/CardList";

const AuthorGenre = () => {
  const { authorID } = useParams<{ authorID: string }>();
  const location = useLocation();
  const books = location.state as Book[];
  const decodedAuthorName = authorID ? decodeURIComponent(authorID).split('-').join(' ') : 'Unknown Author';

  console.log('Author landing page, testing books: ', books);

  return (
    <div className="bk_lib flex flex-col items-center\ px-5 antialiased md:px-1 md:ml-24 h-screen pt-28">
      <h1 className="text-left">Books by {decodedAuthorName}</h1>

      <div className="flex flex-row relative w-full max-w-7xl justify-between items-center text-left text-white border-b-2 border-solid border-zinc-700">
        {/* Number of total items in view  */}
        <div className="mt-1">{books?.length || 0} volumes</div>
      </div>

      { books && books.length > 0 && <CardList books={books} /> }
    </div>
  )
}

export default AuthorGenre;
