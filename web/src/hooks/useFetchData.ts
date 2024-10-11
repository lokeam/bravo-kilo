import { useQuery, useQueryClient, QueryKey, UseQueryOptions, QueryFunction } from '@tanstack/react-query';

type FetchFunction<T, Q> = Q extends undefined ? never : (query: Q) => Promise<T>;

// Type definitions to improve type safety + flexibility
interface UseFetchDataOptions<T, Q> extends Omit<UseQueryOptions<T, Error, T, QueryKey>, 'queryKey' | 'queryFn'> {
  queryKey: QueryKey;
  fetchFunction: FetchFunction<T, Q>;
  query?: Q;
  enabled?: boolean;
  staleTime?: number;
  gcTime?: number;
  retry?: number | boolean;
}

// Type guard function
function isUseFetchDataOptions<T, Q>(obj: any): obj is UseFetchDataOptions<T, Q> {
  return typeof obj === 'object' && 'queryKey' in obj && 'fetchFunction' in obj;
}

// Fn signature accepts either original params or single options obj
const useFetchData = <T, Q>(
  queryKeyOrOptions: QueryKey | UseFetchDataOptions<T, Q>,
  fetchFunctionOrUndefined?: FetchFunction<T, Q>,
  queryOrUndefined?: Q,
  enabledOrUndefined?: boolean
) => {
  // API branching - logic to determine if new or old API used
  if (Array.isArray(queryKeyOrOptions) || typeof queryKeyOrOptions === 'string') {
    // Old API
    const queryKey = queryKeyOrOptions;
    const fetchFunction = fetchFunctionOrUndefined!;
    const query = queryOrUndefined;
    const enabled = enabledOrUndefined ?? true;

    return useQuery<T, Error, T, QueryKey>({
      queryKey: Array.isArray(queryKey) ? queryKey : [queryKey],
      queryFn: () => {
        if (query !== undefined) {
          return fetchFunction(query);
        }
        return Promise.reject(new Error('Query parameter is undefined'));
      },
      enabled: enabled && query !== undefined,
      staleTime: 1000 * 60 * 5,
      gcTime: 1000 * 60 * 15,
      retry: 3,
    });
  } else if (isUseFetchDataOptions<T, Q>(queryKeyOrOptions)) {
    // New API
    const {
      queryKey,
      fetchFunction,
      query,
      enabled = true,
      staleTime = 1000 * 60 * 5,
      gcTime = 1000 * 60 * 15,
      retry = 3,
      ...options
    } = queryKeyOrOptions;

    const fullQueryKey = Array.isArray(queryKey) ? queryKey : [queryKey];
    if (query !== undefined && query !== '') {
      fullQueryKey.push(query as any);
    }

    const queryFn: QueryFunction<T, QueryKey> = () => {
      if (query !== undefined) {
        return fetchFunction(query);
      }
      return Promise.reject(new Error('Query parameter is undefined'));
    };

    return useQuery<T, Error, T, QueryKey>({
      queryKey: fullQueryKey,
      queryFn,
      enabled: enabled && query !== undefined,
      staleTime,
      gcTime,
      retry,
      ...options,
    });
  } else {
    throw new Error('Invalid arguments passed to useFetchData');
  }
};

// Utility fns:
export const invalidateQueries = (queryClient: ReturnType<typeof useQueryClient>, queryKey: string | string[]) => {
  queryClient.invalidateQueries({ queryKey: Array.isArray(queryKey) ? queryKey : [queryKey] });
};

export const setQueryData = <T>(
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: string | string[],
  updater: T | ((oldData: T | undefined) => T)
) => {
  queryClient.setQueryData<T>(Array.isArray(queryKey) ? queryKey : [queryKey], updater);
};

export default useFetchData;
