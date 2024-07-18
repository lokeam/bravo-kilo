import { Outlet } from "react-router-dom";
import TopNavigation from "../components/TopNav/TopNav";
import SideNavigation from "../components/SideNav/SideNavigation";

import { useAuth } from "../components/AuthContext";

const Library = () => {
  const { logout } = useAuth();

  return (
    <div className="bk_lib">
      <TopNavigation />
      <SideNavigation />

      <h1>Library</h1>

      <button onClick={logout}>Sign out of your Kilo Bravo account</button>

      <Outlet />
    </div>
  )
}

export default Library;
