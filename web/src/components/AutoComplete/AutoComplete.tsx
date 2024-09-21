import React, { useEffect, useState, useRef, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import useBookSearch from '../../hooks/useBookSearch';
import useSearchStore from '../../store/useSearchStore';
import { IoClose, IoSearchOutline } from 'react-icons/io5';

function AutoComplete() {
  const [query, setQuery] = useState('');
  const [searchParams, setSearchParams] = useSearchParams();
  const [highlightedIndex, setHighlightedIndex] = useState<number | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const { addSearchHistory, getFilteredSearchHistory } = useSearchStore();

  const suggestions = Object.keys(getFilteredSearchHistory()).filter((key) =>
    key.toLowerCase().startsWith(query.toLowerCase())
  );

  // Use hook to fetch data when the query is submitted
  const { data } = useBookSearch(searchParams.get('query') || '');

  // Handlers - Clicking outside suggestions list
  const handleClickOutside = useCallback((event: MouseEvent) => {
    if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
      setHighlightedIndex(null);
    }
  }, []);

  // Handlers - Setting focus highlight
  const handleFocus = useCallback(() => {
    if (query.length > 0 && suggestions.length > 0) {
      setHighlightedIndex(0);
    }
  }, [query, suggestions]);

  // Handlers - Setting blur
  const handleBlur = useCallback((event: React.FocusEvent<HTMLDivElement>) => {
    if (!containerRef.current?.contains(event.relatedTarget as Node)) {
      setHighlightedIndex(null);
    }
  }, []);

  // Handlers - Input change
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setQuery(e.target.value);
    setHighlightedIndex(e.target.value.length > 0 && suggestions.length > 0 ? 0 : null);
  };

  // Handlers - Keyboard navigation
  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (highlightedIndex !== null) {
      if (event.key === 'ArrowDown') {
        setHighlightedIndex((prevIndex) =>
          prevIndex === null || prevIndex === suggestions.length - 1 ? 0 : prevIndex + 1
        );
      } else if (event.key === 'ArrowUp') {
        setHighlightedIndex((prevIndex) =>
          prevIndex === null || prevIndex === 0 ? suggestions.length - 1 : prevIndex - 1
        );
      } else if (event.key === 'Enter') {
        if (highlightedIndex !== null) {
          handleSuggestionClick(suggestions[highlightedIndex]);
        } else {
          // Submit the form if no suggestion is highlighted
          event.preventDefault();
          handleSubmit(event);
        }
      } else if (event.key === 'Escape') {
        setHighlightedIndex(null);
      }
    }
  };

  // Handlers - Suggestion click
  const handleSuggestionClick = (suggestion: string) => {
    setQuery(suggestion);
    setHighlightedIndex(null);
    setSearchParams({ query: suggestion });
    inputRef.current?.focus();
  };

  // Handlers - Submitting query
  const handleSubmit = useCallback((event: React.FormEvent<HTMLFormElement> | React.KeyboardEvent<HTMLInputElement>) => {
    if (event) event.preventDefault();

    const trimmedQuery = query.trim();

    if (trimmedQuery !== '') {
      setSearchParams({ query: trimmedQuery });
      setHighlightedIndex(null);
    }
  }, [query, setSearchParams]);

  // Handlers - Clear input
  const handleClearInput = () => {
    setQuery('');
    setHighlightedIndex(null);
    inputRef.current?.focus();
  };

  useEffect(() => {
    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [handleClickOutside]);

  useEffect(() => {
    if (data && data.books && searchParams.get('query')) {
      console.log('Autocomplete.tsx: Search results: ', data);
      addSearchHistory(searchParams.get('query')!, data.books);
    }
  }, [data, searchParams, addSearchHistory]);



  return (
    <div
    className={`
      autocomplete-container w-full border rounded-lg dark:bg-maastricht border-gray-600
      ${
        highlightedIndex !== null && suggestions.length > 0 ? 'border-t border-x rounded-b-none border-b-black' : ''
      }
      `}
      ref={containerRef}
      onFocus={handleFocus}
      onBlur={handleBlur}
    >
      <form
        onSubmit={handleSubmit}
        className="searchbox_container w-full flex relative">
        <div
          className="searchbox__clear_search_field w-full border-none bg-white flex outline-none rounded-full flex-row items-center justify-center dark:bg-maastricht"
        >
          <button
            className="searchbox__clear_search_field border-none rounded-lg bg-white inline-flex outline-none flex-row items-center justify-center dark:bg-maastricht"
            type="submit"
          >
            <IoSearchOutline size={24} />
          </button>
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={handleInputChange}
            onKeyDown={handleKeyDown}
            className={`autocomplete-input w-full bg-white text-az-white font-bold outline-none block  pl-4 p-2.5 placeholder:polo-blue placeholder:font-bold dark:bg-maastricht`}
            placeholder="Add a book via Search"
          />
          <button
            className="inline-flex border-none outline-none rounded-lg flex-row items-center justify-center bg-maastricht"
            onClick={handleClearInput}
            type="button"
          >
            <IoClose size={18} />
          </button>
        </div>
      </form>
      { highlightedIndex !== null && suggestions.length > 0 && (
        <ul
          className={`autocomplete-suggestions ${
            highlightedIndex !== null ? 'rounded-t-none border-t-0' : ''
          } absolute left-0 box-border bg-maastricht border border-gray-600 rounded w-full`}
        >
          {suggestions.map((suggestion, index) => (
            <li
              key={index}
              onClick={() => handleSuggestionClick(suggestion)}
              className={`cursor-pointer text-az-white border-transparent rounded-t-none rounded-lg font-bold text-left pl-20 p-2.5 hover:bg-dark-ebony ${
                index === highlightedIndex ? 'bg-dark-ebony' : ''
              }`}
            >
              {suggestion}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

export default AutoComplete;
