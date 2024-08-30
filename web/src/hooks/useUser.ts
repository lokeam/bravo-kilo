import { useAuth } from '../components/AuthContext';

export const useUser = () => {
  const { user, loading, isAuthenticated } = useAuth();

  return {
    data: user,
    isLoading: loading,
    isAuthenticated,
    isError: !loading && !isAuthenticated,
  };
};