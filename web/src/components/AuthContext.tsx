import { createContext, useContext, useState, ReactNode, useEffect } from 'react';
import { useQuery, QueryClient, QueryClientProvider } from '@tanstack/react-query';
import axios from 'axios';

export interface User {
  id: number;
  email?: string;
  firstName?: string;
  lastName?: string;
  picture?: string;
  createdAt?: string;
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
  const { data } = await axios.get(`${import.meta.env.VITE_API_ENDPOINT}/auth/token/verify`, { withCredentials: true });
  console.log('AuthContext - fetch user data: ', data);

  return data.user;
}

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(true);

  const { data, isLoading, isError } = useQuery({
    queryKey: ['user'],
    queryFn: fetchUser,
    retry: false,
    // Always fetch user data when AuthProvider mounts
    enabled: true,
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
      setLoading(false);
      setIsAuthenticated(false);
    }
  }, [isError]);

  const login = () => {
    window.location.href = `${import.meta.env.VITE_API_ENDPOINT}/auth/google/signin`;
  };

  const logout = async () => {
    await axios.post(`${import.meta.env.VITE_API_ENDPOINT}/auth/signout`, {}, { withCredentials: true });
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
    return <div>Loading...</div>;
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
