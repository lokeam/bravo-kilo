import React from 'react';

interface SideNavListProps {
  children: React.ReactNode;
}

function SideNavList({ children }: SideNavListProps) {
  return <>{children}</>;
}

export default SideNavList;
