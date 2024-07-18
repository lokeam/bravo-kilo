import React from 'react';

interface SideNavListProps {
  children: React.ReactNode;
}

export const SideNavList = ({ children }: SideNavListProps) => {
  return <div>{children}</div>;
};
