import { useParams, useLocation, useSearchParams } from "react-router-dom";
import { Book } from "./Library";
import CardList from "../components/CardList/CardList";

const AuthorGenre = () => {
  const { authorID = "" } = useParams<{ authorID: string }>();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const isAuthorPage = searchParams && searchParams.get('page') === 'author';
  const books = location.state as Book[];
  const decodedPageID = authorID ? decodeURIComponent(authorID).split('-').join(' ') : 'Unknown Author';

  console.log('decoded identifier: ', decodedPageID);
  console.log('testing isAuthorPage: ', isAuthorPage );

  console.log('Author landing page, testing books: ', books);
  return (
    <div className="bk_lib flex flex-col items-center px-5 antialiased mdTablet:pr-5 mdTablet:ml-24 h-screen pt-28">
      <h1 className="text-left">
        { isAuthorPage ? `Books by ${decodedPageID}` : decodedPageID }
      </h1>

      <div className="flex flex-row relative w-full max-w-7xl justify-between items-center text-left text-white border-b-2 border-solid border-zinc-700">
        {/* Number of total items in view  */}
        <div className="mt-1">{books?.length || 0} volumes</div>
      </div>

      { books && books.length > 0 && <CardList books={books} /> }
    </div>
  )
}

export default AuthorGenre;
