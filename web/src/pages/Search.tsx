import { useSearchParams } from 'react-router-dom';
import CardList from '../components/CardList/CardList';
import Loading from '../components/Loading/Loading';
import useSearchStore from '../store/useSearchStore';
import useBookSearch from '../hooks/useBookSearch';
import { TbWorldSearch } from "react-icons/tb";
import PageWithErrorBoundary from '../components/ErrorMessages/PageWithErrorBoundary';

function Search() {
  const [searchParams] = useSearchParams();
  const query = searchParams.get('query') || '';
  const { searchHistory } = useSearchStore();
  const { isLoading } = useBookSearch(query);

  const searchEntry = searchHistory[query];
  const books = searchEntry ? searchEntry.results : [];

  const renderContent = () => {
    if (!query) {
      return (
        <div className="text-center pt-28">
          <TbWorldSearch
            className="mx-auto text-6xl mb-4 text-charcoal dark:text-white"
            size={38}
          />
          <p className="text-xl text-black dark:text-white">Search for a book to get started!</p>
        </div>
      );
    }

    if (isLoading) {
      return (
        <div className="bk_searchflex flex-col items-center px-5 antialiased mdTablet:pl-1 pr-5 mdTablet:ml-24 h-screen pt-28 ">
          <Loading />
        </div>
      );
    }

    if (books.length === 0) {
      return (
        <p className="text-center">We couldn't find any results found for {query}</p>
      );
    }

    return (
      <CardList
        books={books}
        isSearchPage
      />
    );
  }

  return (
    <PageWithErrorBoundary fallbackMessage="Error loading search page">
      <div className="bk_search min-h-screen bg-cover bg-white flex flex-col items-center px-5 antialiased mdTablet:pl-1 pr-5 mdTablet:ml-24 dark:bg-black">
        { renderContent() }
      </div>
    </PageWithErrorBoundary>
  );
}

export default Search;
