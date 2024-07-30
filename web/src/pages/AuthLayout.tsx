
import { Outlet } from 'react-router-dom';
import TopNavigation from '../components/TopNav/TopNav';
import SideNavigation from '../components/SideNav/SideNavigation';

const AuthenticatedLayout = () => {
  return (
    <div className="authenticated-layout">
      <TopNavigation />
      <SideNavigation />
      <div className="content">
        <Outlet />
      </div>
    </div>
  );
};

export default AuthenticatedLayout;
