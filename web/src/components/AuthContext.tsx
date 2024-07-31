import { createContext, useContext, useState, ReactNode, useEffect } from 'react';
import { useQuery, QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { verifyUserToken, signOutUser } from '../service/apiClient.service';

export interface User {
  id?: number;
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

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(true);

  const { data, isLoading, isError } = useQuery({
    queryKey: ['user'],
    queryFn: verifyUserToken,
    retry: false,
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
    await signOutUser();
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
