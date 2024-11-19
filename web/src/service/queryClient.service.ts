import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      gcTime: 1000 * 60 * 10, // 10 minutes
      retry: (failureCount, error) => {
        if (error instanceof Error && error.message.includes('Network Error')) {
          return failureCount < 3;
        }
        return false;
      },
      refetchOnWindowFocus: false,
    },
  },
});
