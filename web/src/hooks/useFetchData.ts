import { useQuery } from '@tanstack/react-query';

type FetchFunction<T, Q> = (query: Q) => Promise<T>;

const useFetchData = <T, Q>(queryKey: string, fetchFunction: FetchFunction<T, Q>, query: Q, enabled: boolean = true) => {
  return useQuery({
    queryKey: [queryKey, query],
    queryFn: () => fetchFunction(query),
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 5,
    enabled: enabled,
  });
};

export default useFetchData;
