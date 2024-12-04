import { createContext, useContext, useState, ReactNode, useEffect } from 'react';
import { useQuery, QueryClient, QueryClientProvider } from '@tanstack/react-query';
import Loading from './Loading/Loading';
import apiClient from '../service/apiClient.service';
import { signOutUser } from '../service/apiClient.service';
import { useLocation } from 'react-router-dom';

export interface User {
  id: number;
  email?: string;
  firstName?: string;
  lastName?: string;
  picture?: string;
  created_at?: string;
  updatedAt?: string;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  login: () => void;
  logout: () => void;
  loading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const fetchUser = async() => {

  try {
    const { data } = await apiClient.get('/auth/token/verify');
    //const { data } = await axios.get(`${import.meta.env.VITE_API_ENDPOINT}/auth/token/verify`, { withCredentials: true });
    console.log('AuthContext - fetch user data: ', data);
    return data.user;
  } catch(error) {
    // make sure error is caught by useQuery
    console.error('Error fetching user: ', error);
    throw error;
  }
}

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(true);

  const location = useLocation();

  const { data, isLoading, isError } = useQuery({
    queryKey: ['user'],
    queryFn: fetchUser,
    retry: false,
    // Always fetch user data when AuthProvider mounts, except on login page
    enabled: location.pathname !== '/login',
  });

  useEffect(() => {
    if (data) {
      console.log('User data available in useEffect:', data);
      setUser(data);
      setIsAuthenticated(true);
    } else {
      console.log('No user data in useEffect');
      setIsAuthenticated(false);
    }
    setLoading(isLoading);
  }, [data, isLoading]);

  useEffect(() => {
    if (isError) {
      setUser(null);
      setLoading(false);
      setIsAuthenticated(false);
    }
  }, [isError]);

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!loading && !isAuthenticated && location.pathname !== '/login') {
      window.location.href = '/login';
    }
  }, [loading, isAuthenticated, location.pathname])

  const login = () => {
    window.location.href = `${import.meta.env.VITE_API_ENDPOINT}/auth/google/signin`;
  };

  const logout = async () => {
    //await axios.post(`${import.meta.env.VITE_API_ENDPOINT}/auth/signout`, {}, { withCredentials: true });
    signOutUser();
    setUser(null);
    setIsAuthenticated(false);
    window.location.href = "/login";
  };

  const value = {
    user,
    isAuthenticated,
    login,
    logout,
    loading,
  };

  console.log('AuthContext value:', value);

   // Stop children from rendering until user data is ready
  if (loading) {
    return <Loading />;
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used within an AuthProvider");
  return context;
};

const queryClient = new QueryClient();

export const AppProvider = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={queryClient}>
    <AuthProvider>{children}</AuthProvider>
  </QueryClientProvider>
);