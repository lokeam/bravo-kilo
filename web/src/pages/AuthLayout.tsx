import { useMemo } from 'react';
import { Outlet } from 'react-router-dom';
import TopNavigation from '../components/TopNav/TopNav';
import SideNavigation from '../components/SideNav/SideNavigation';
import OfflineBanner from '../components/ErrorMessages/OfflineBanner';
import Snackbar from '../components/Snackbar/Snackbar';
import useStore from '../store/useStore';

const AuthenticatedLayout = () => {
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

  return (
    <div className="authenticated-layout h-dvh bg-white dark:bg-black">
      <TopNavigation />
      <SideNavigation />
      <div className="content pt-[67px]">
        <OfflineBanner />
        {memoizedSnackbar}
        <Outlet />
      </div>
    </div>
  );
};

export default AuthenticatedLayout;
