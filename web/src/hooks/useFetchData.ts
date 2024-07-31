import { useQuery } from '@tanstack/react-query';

type FetchFunction<T> = (userID: number) => Promise<T>;

const useFetchData = <T>(queryKey: string, fetchFunction: FetchFunction<T>, userID: number) => {
  return useQuery({
    queryKey: [queryKey, userID],
    queryFn: () => fetchFunction(userID),
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 5,
  });
};

export default useFetchData;
