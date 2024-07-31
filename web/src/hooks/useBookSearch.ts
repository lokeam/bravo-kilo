import { useQuery } from '@tanstack/react-query';
import { searchBookAPI } from '../service/apiClient.service';

const useBookSearch = (query: string) => {
  return useQuery({
    queryKey: ['bookSearch', query],
    queryFn: () => searchBookAPI(query),
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 5,
    enabled: !!query,  // Ensures the query only runs if the query string is not empty
  });
};

export default useBookSearch;
