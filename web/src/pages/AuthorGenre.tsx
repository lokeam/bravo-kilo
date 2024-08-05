import { useParams, useLocation } from "react-router-dom";
import { Book } from "./Library";

const AuthorGenre = () => {
  const { authorID } = useParams<{ authorID: string }>();
  const location = useLocation();
  const books = location.state as Book[];

  console.log('Author landing page, testing books: ', books);

  return (
    <div className="bk_lib flex flex-col items-center place-content-around px-5 antialiased md:px-1 md:ml-24 h-screen pt-28">
      <h1>Books by {authorID ? decodeURIComponent(authorID) : 'Unknown Author'}</h1>
      <ul>
      </ul>
    </div>
  )
}

export default AuthorGenre;
