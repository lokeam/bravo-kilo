import React from 'react';
import { IconType } from 'react-icons';

interface SideNavListItemProps {
  href: string;
  icon?: IconType;
  children: React.ReactNode;
}

export const SideNavListItem = ({ href, icon: Icon, children }: SideNavListItemProps) => {
  return (
    <a href={href} className="flex items-center p-4 hover:bg-gray-700">
      { Icon ? (<Icon className="w-5 h-5 mr-2" />) : null }
      {children}
    </a>
  );
};
