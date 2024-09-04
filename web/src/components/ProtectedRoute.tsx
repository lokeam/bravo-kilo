import { Navigate } from 'react-router-dom';
import { useAuth } from './AuthContext';

interface ProtectedRouteProps {
  children: JSX.Element;
}

function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isAuthenticated, loading } = useAuth();

  console.log('ProtectedRoute: isAuthenticated', isAuthenticated);
  console.log('ProtectedRoute: loading', loading);

  if (loading) {
    return <div>Loading...</div>;
  }

  if (!isAuthenticated) {
    console.log('Redirecting to /login');
    return <Navigate to="/login" />;
  }

  return children;
}

export default ProtectedRoute;
