import { useQuery } from '@tanstack/react-query';
import { searchBookAPI } from '../service/apiClient.service';

// queryFn: () => searchBookAPI(query),

const useBookSearch = (query: string) => {
  const enabled = query.trim().length > 0;
  return useQuery({
    queryKey: ['bookSearch', query],
    queryFn: async () => {
      const books = await searchBookAPI(query);
      console.log('useBookSearch.ts: response:', books); // Log the books data
      return books;
    },
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 5,
    enabled,  // Ensures the query only runs if the query string is not empty
  });
};

export default useBookSearch;
