import React from 'react';
import { useSearchParams } from 'react-router-dom';
import CardList from '../components/CardList/CardList';
import useSearchStore from '../store/useSearchStore';

const Search = () => {
  const [searchParams] = useSearchParams();
  const query = searchParams.get('query') || '';
  const { searchHistory } = useSearchStore();

  // Retrieve the latest search results from the store using the query from the URL
  const searchEntry = searchHistory[query];
  const books = searchEntry ? searchEntry.results : [];

  console.log('Search Page');
  console.log('Search Page grabbing searchHistory from useSearch Store: ', searchEntry);
  console.log('Search Page, raw searchParams: ', searchParams);
  console.log('Search Page, getting query Search Params: ', query);


  return (
    <div className="bk_lib flex flex-col px-5 antialiased md:px-1 md:ml-24 h-screen pt-20">
      {books && books.length > 0 ? (
        <CardList books={books} />
      ) : (
        <p>No results found.</p>
      )}
    </div>
  );
};

export default Search;
