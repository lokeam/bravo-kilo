import React, { forwardRef } from 'react';
import { NavLink } from 'react-router-dom';
import { IconType } from 'react-icons';

interface SideNavListItemProps {
  pageRoute: string;
  icon?: IconType;
  children: React.ReactNode;
}

// Use forwardRef to pass the ref to the NavLink
const SideNavListItem = forwardRef<HTMLAnchorElement, SideNavListItemProps>(
  ({ pageRoute, icon: Icon, children }, ref) => {
    return (
      <NavLink
        ref={ref} // Attach the forwarded ref to the NavLink component
        className={`
          flex flex-col w-16 md:w-auto items-center text-xs rounded text-center text-charcoal py-1 md:py-2 my-3 hover:text-az-white hover:bg-gray-700 dark:text-cadet-gray focus:bg-gray-700 focus-within:bg-gray-700 focus:text-white focus-within:text-white dark:focus:text-white dark:focus-within:text-white
          `}
        end
        to={pageRoute}
      >
        {({ isActive }) => (
          <>
            {Icon ? (
              <Icon className={`w-5 h-5 mb-1 ${isActive ? '' : ''}`} />
            ) : null}
            <span className={`${isActive ? 'font-bold' : ''}`}>{children}</span>
          </>
        )}
      </NavLink>
    );
  }
);

export default SideNavListItem;
