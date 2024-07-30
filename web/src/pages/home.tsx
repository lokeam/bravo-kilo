
import useFetchBooks from "../hooks/useFetchBooks";
import { useLocation } from "react-router-dom";

import Bookshelf from "../components/Bookshelf/Bookshelf";

const Home = () => {
  const { search } = useLocation();
  const query = new URLSearchParams(search);
  const userID = parseInt(query.get('userID') || '0', 10);
  const { data: books, isLoading, isError } = useFetchBooks(userID);

  console.log('dashboard home - books: ', books);

  return (
    <div className="bk_home flex flex-col px-5 antialiased md:px-1 md:ml-24 h-screen pt-40">
      <Bookshelf category="Recently updated" books={books || []} isLoading={isLoading} />
      <button>Add a shelf</button>
    </div>
  )
}

export default Home;
