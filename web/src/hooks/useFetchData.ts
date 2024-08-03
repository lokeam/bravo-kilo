import { useQuery } from '@tanstack/react-query';

type FetchFunction<T, Q> = (query: Q) => Promise<T>;

const useFetchData = <T, Q>(queryKey: string, fetchFunction: FetchFunction<T, Q>, query: Q, enabled: boolean) => {
  return useQuery({
    queryKey: [queryKey, query],
    queryFn: () => fetchFunction(query),
    enabled,
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 15,
    retry: 3,
  });
};

export default useFetchData;
