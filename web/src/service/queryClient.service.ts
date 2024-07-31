import { QueryClient } from '@tanstack/react-query';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: true,
      refetchInterval: 1000 * 60 * 5,
      staleTime: 1000 * 60 * 5.
    },
  },
});

export default queryClient;
