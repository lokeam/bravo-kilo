import React from 'react';
import { IconType } from 'react-icons';

interface SideNavListItemProps {
  href: string;
  icon?: IconType;
  children: React.ReactNode;
}

export const SideNavListItem = ({ href, icon: Icon, children }: SideNavListItemProps) => {
  return (
    <a href={href} className="flex flex-col w-16 md:w-auto items-center text-xs rounded text-center text-cadet-gray py-1 md:py-2 my-3 hover:text-az-white hover:bg-gray-700">
      { Icon ? (<Icon className="w-5 h-5 mb-1" />) : null }
      <span>{children}</span>
    </a>
  );
};
