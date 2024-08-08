import React, { useEffect, useState, useRef, useCallback } from 'react';
import useSearchStore from '../../store/useSearchStore';
import { SearchResult } from '../../store/useSearchStore';

interface AutoCompleteProps {
  onSubmit: (query: string) => void;
}

const AutoComplete: React.FC<AutoCompleteProps> = ({ onSubmit }) => {
  const [query, setQuery] = useState('');
  const [highlightedIndex, setHighlightedIndex] = useState<number | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const { addSearchHistory, getFilteredSearchHistory } = useSearchStore();

  const suggestions = Object.keys(getFilteredSearchHistory()).filter((key) =>
    key.toLowerCase().startsWith(query.toLowerCase())
  );

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
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (highlightedIndex !== null) {
      if (e.key === 'ArrowDown') {
        setHighlightedIndex((prevIndex) =>
          prevIndex === null || prevIndex === suggestions.length - 1 ? 0 : prevIndex + 1
        );
      } else if (e.key === 'ArrowUp') {
        setHighlightedIndex((prevIndex) =>
          prevIndex === null || prevIndex === 0 ? suggestions.length - 1 : prevIndex - 1
        );
      } else if (e.key === 'Enter') {
        if (highlightedIndex !== null) {
          handleSuggestionClick(suggestions[highlightedIndex]);
        }
      } else if (e.key === 'Escape') {
        setHighlightedIndex(null);
      }
    }
  };

  // Handlers - Input change
  const handleSuggestionClick = (suggestion: string) => {
    setQuery(suggestion);
    setHighlightedIndex(null);
    inputRef.current?.focus();
  };

  // Handlers - Submitting query
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const trimmedQuery = query.trim();

    if (trimmedQuery !== '') {
      const fetchedResults: SearchResult[] = [];
      addSearchHistory(trimmedQuery, fetchedResults);
      onSubmit(trimmedQuery);
      setHighlightedIndex(null);
    }
  };

  useEffect(() => {
    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [handleClickOutside]);

  return (
    <div
      className="autocomplete-container w-full"
      ref={containerRef}
      onFocus={handleFocus}
      onBlur={handleBlur}
    >
      <form onSubmit={handleSubmit} className="w-full relative">
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          className={`autocomplete-input rounded ${
            highlightedIndex !== null ? 'border-t border-x rounded-b-none' : 'border'
          } bg-maastricht border-gray-600 text-az-white font-bold focus:ring-primary-500 focus:border-primary-500 block w-full pl-9 p-2.5 placeholder:polo-blue placeholder:font-bold`}
          placeholder="Search for a book or author"
        />
      </form>
      {highlightedIndex !== null && (
        <ul
          className={`autocomplete-suggestions ${
            highlightedIndex !== null ? 'rounded-t-none' : ''
          } absolute bg-maastricht border border-gray-600 rounded w-full`}
        >
          {suggestions.map((suggestion, index) => (
            <li
              key={index}
              onClick={() => handleSuggestionClick(suggestion)}
              className={`cursor-pointer text-az-white font-bold text-left pl-9 p-2.5 hover:bg-dark-ebony ${
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
};

export default AutoComplete;
