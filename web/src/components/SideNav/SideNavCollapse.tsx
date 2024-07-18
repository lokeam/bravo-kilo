import React, { useState, useEffect } from 'react';
import Cookies from 'js-cookie';
import { IconType } from 'react-icons';
import { MdExpandMore, MdExpandLess } from 'react-icons/md';

interface SideNavCollapseProps {
  label: string;
  icon: IconType;
  children: React.ReactNode;
  collapseKey: string;
}

export const SideNavCollapse = ({ label, icon: Icon, children, collapseKey }: SideNavCollapseProps) => {
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    const cookieState = Cookies.get(collapseKey);
    setIsOpen(cookieState === 'true');
  }, [collapseKey]);

  const toggleCollapse = () => {
    const nextIsOpen = !isOpen;
    setIsOpen(nextIsOpen);
    Cookies.set(collapseKey, nextIsOpen.toString(), { expires: 7 });
  };

  return (
    <div>
      <button
        onClick={toggleCollapse}
        className="group flex w-full items-center justify-between p-4 text-left hover:bg-gray-700"
        aria-expanded={isOpen}
        aria-controls={`${collapseKey}-content`}
      >
        <div className="flex items-center">
          <Icon className="w-5 h-5 mr-2" />
          <span>{label}</span>
        </div>
        {isOpen ? <MdExpandLess className="w-5 h-5" /> : <MdExpandMore className="w-5 h-5" />}
      </button>
      {isOpen && <div id={`${collapseKey}-content`} className="ml-4">{children}</div>}
    </div>
  );
};
