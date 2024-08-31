import { useSearchParams } from 'react-router-dom';
import CardList from '../components/CardList/CardList';
import Loading from '../components/Loading/Loading';
import useSearchStore from '../store/useSearchStore';
import useBookSearch from '../hooks/useBookSearch';
import { TbWorldSearch } from "react-icons/tb";

function Search() {
  const [searchParams] = useSearchParams();
  const query = searchParams.get('query') || '';
  const { searchHistory } = useSearchStore();
  const { isLoading, error } = useBookSearch(query);

  const searchEntry = searchHistory[query];
  const books = searchEntry ? searchEntry.results : [];

  const renderContent = () => {
    if (!query) {
      return (
        <div className="text-center">
          <TbWorldSearch
            className="mx-auto text-6xl mb-4"
            size={30}
          />
          <p className="text-xl">Search for a book to get started!</p>
        </div>
      );
    }

    if (isLoading) return <Loading />;

    if (error) {
      return (
        <div className="text-center text-red-500">
          <p>Something happened. Don't worry, we're working on it. Please try again later</p>
          <p>Error: {error.message}</p>
        </div>
      );
    }

    if (books.length === 0) {
      return (
        <p className="text-center">No results found for {query}</p>
      );
    }

    return (
      <CardList
        books={books}
        isSearchPage
      />
    )
  }

  return (
    <div className="bk_search flex flex-col items-center place-content-around px-5 antialiased mdTablet:pr-5 mdTablet:ml-24 h-screen pt-28">
      {renderContent()}
    </div>
  );
}

export default Search;
