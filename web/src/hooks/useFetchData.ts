import { useQuery } from '@tanstack/react-query';

type FetchFunction<T, Q> = (query: Q | undefined) => Promise<T>;

const useFetchData = <T, Q>(queryKey: string, fetchFunction: FetchFunction<T, Q>, query: Q | undefined, enabled: boolean = true) => {
  return useQuery({
    queryKey: query ? [queryKey, query] : [queryKey],
    queryFn: () => {
      if (query !== undefined) {
        return fetchFunction(query);
      }
      return Promise.reject('Query parameter is undefined');
    },
    enabled: enabled && query !== undefined, // Enable query only if `query` is defined
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 15,
    retry: 3,
  });
};

export default useFetchData;
