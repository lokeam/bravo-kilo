import { SideNavWrapper } from './SideNavWrapper';
import { SideNavLogo } from './SideNavLogo';
import { SideNavList } from './SideNavList';
import { SideNavListItem } from './SideNavListItem';
import { SideNavCollapse } from './SideNavCollapse';

// Icons and Logos
import { GoHomeFill } from "react-icons/go";
import { BiLibrary } from "react-icons/bi";

import { MdOutlineGridView } from 'react-icons/md';
import { HiOutlineCollection } from 'react-icons/hi';
import { MdCategory } from "react-icons/md";
import { TbEdit } from "react-icons/tb";

import bkLogo from '../../assets/tk_icon.webp'



export default function SideNavigation() {

  return (
    <SideNavWrapper className="" ariaLabel="Main navigation">

      <SideNavList>
        <SideNavListItem href="#" icon={GoHomeFill}>Home</SideNavListItem>
        <SideNavListItem href="#" icon={BiLibrary}>Library</SideNavListItem>
        <SideNavListItem href="#" icon={MdCategory}>Categories</SideNavListItem>
        <SideNavListItem href="#" icon={TbEdit}>Add</SideNavListItem>

      </SideNavList>
    </SideNavWrapper>
  );
}
