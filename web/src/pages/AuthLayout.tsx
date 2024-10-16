import { useMemo } from 'react';
import { useAuth } from '../components/AuthContext';
import { Navigate } from 'react-router-dom';
import { Outlet } from 'react-router-dom';
import TopNavigation from '../components/TopNav/TopNav';
import SideNavigation from '../components/SideNav/SideNavigation';
import OfflineBanner from '../components/ErrorMessages/OfflineBanner';
import PageTransition from '../components/PageTransition';
import Snackbar from '../components/Snackbar/Snackbar';
import useStore from '../store/useStore';

const AuthenticatedLayout = () => {
  const { isAuthenticated } = useAuth();
  const {
    snackbarMessage,
    snackbarOpen,
    snackbarVariant,
    hideSnackbar,
  } = useStore();

  const memoizedSnackbar = useMemo(() => (
    <Snackbar
      message={snackbarMessage || ''}
      open={snackbarOpen}
      variant={snackbarVariant || 'added'}
      onClose={hideSnackbar}
    />
  ), [snackbarMessage, snackbarOpen, snackbarVariant, hideSnackbar]);

  if (!isAuthenticated) {
    return <Navigate to="/login" />;
  }

  return (
    <div className="authenticated-layout h-dvh bg-white dark:bg-black">
      <TopNavigation />
      <SideNavigation />
      <div className="content pt-[67px] bg-white dark:bg-black">
        <OfflineBanner />
        {memoizedSnackbar}
        <PageTransition>
          <Outlet />
        </PageTransition>
      </div>
    </div>
  );
};

export default AuthenticatedLayout;
