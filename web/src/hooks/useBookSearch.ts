import { useQuery } from '@tanstack/react-query';
import { searchBookAPI } from '../service/apiClient.service';

const useBookSearch = (query: string) => {
  const enabled = query.trim().length > 0;
  return useQuery({
    queryKey: ['bookSearch', query],
    queryFn: () => searchBookAPI(query),
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 5,
    enabled,  // Ensures the query only runs if the query string is not empty
  });
};

export default useBookSearch;
