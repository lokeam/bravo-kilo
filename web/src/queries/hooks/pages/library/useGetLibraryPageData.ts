import { useQuery } from '@tanstack/react-query';
import { queryKeys } from '../../../constants/queryKeys';
import { fetchLibraryPageData } from '../../../../service/apiClient.service';
import { AxiosError } from 'axios';

interface UseGetLibraryPageDataOptions {
  domain?: string;
  userID: number;
  staleTime?: number;
  gcTime?: number;
  maxRetries?: number;
}

export function useGetLibraryPageData({
  domain = 'books',
  userID,
  staleTime = 5 * 60 * 1000, // 5min
  gcTime = 10 * 60 * 1000,  // 10min
  maxRetries = 3,
}: UseGetLibraryPageDataOptions) {

  if (!userID) {
    console.error('useGetLibraryPageData called without userID');
  }

  return useQuery({
    // Include ID for cache management
    queryKey: [...queryKeys.library.page.byDomain(domain), userID],
    queryFn: async () =>  {
      try {
        // Log query execution
        console.log('Executing library page query:', {
          userID,
          domain,
          queryKey: [...queryKeys.library.page.byDomain(domain), userID]
        });

        const response = await fetchLibraryPageData(userID);
        // Log successful response
        console.log('Library page query successful:', {
          requestId: response.requestId,
          dataPresent: !!response
        });

        return response;
      } catch (error){
        if (error instanceof AxiosError) {
          console.error('Library page query failed:', {
            status: error.response?.status,
            data: error.response?.data,
            message: error.message,
            config: {
              url: error.config?.url,
              params: error.config?.params
            }
          });
        }
        throw error;
      }
    },
    staleTime,
    gcTime,
    retry: (failureCount, error) => {
      if (error instanceof AxiosError) {
        // Log retry attempts
        console.warn(`Retry attempt ${failureCount}:`, {
          status: error.response?.status,
          data: error.response?.data
        });

        if (error.response?.status === 400) {
          console.error('Domain validation failed:', error.response.data);
          return false;
        }
        if (error.response?.status === 401) return false;
        if (error.message.includes('Network Error')) {
          return failureCount < maxRetries;
        }
      }
      return false;
    },
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    select: (data) => {
      // Optionally transform or validate response if needed
      return data;
    },
    enabled: !!userID,
  });
}