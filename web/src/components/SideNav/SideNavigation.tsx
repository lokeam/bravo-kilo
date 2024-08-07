import { SideNavWrapper } from './SideNavWrapper';
import { SideNavList } from './SideNavList';
import { SideNavListItem } from './SideNavListItem';

// Icons and Logos
import { GoHomeFill } from "react-icons/go";
import { BiLibrary } from "react-icons/bi";
import { TbEdit } from "react-icons/tb";
import { IoSearchOutline } from 'react-icons/io5';

export default function SideNavigation() {

  return (
    <SideNavWrapper className="" ariaLabel="Main navigation">
      <SideNavList>
        <SideNavListItem pageRoute="/home" icon={GoHomeFill}>Home</SideNavListItem>
        <SideNavListItem pageRoute="/library" icon={BiLibrary}>Library</SideNavListItem>
        <SideNavListItem pageRoute="/library/books/search" icon={IoSearchOutline}>Search</SideNavListItem>
        <SideNavListItem pageRoute="/library/books/add" icon={TbEdit}>Add</SideNavListItem>
      </SideNavList>
    </SideNavWrapper>
  );
}
