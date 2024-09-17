
import { Outlet } from 'react-router-dom';
import TopNavigation from '../components/TopNav/TopNav';
import SideNavigation from '../components/SideNav/SideNavigation';
import OfflineBanner from '../components/ErrorMessages/OfflineBanner';

const AuthenticatedLayout = () => {
  return (
    <div className="authenticated-layout">
      <TopNavigation />
      <SideNavigation />
      <div className="content pt-[67px]">
        <OfflineBanner />
        <Outlet />
      </div>
    </div>
  );
};

export default AuthenticatedLayout;
