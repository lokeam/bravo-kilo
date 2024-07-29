import { SideNavWrapper } from './SideNavWrapper';
import { SideNavLogo } from './SideNavLogo';
import { SideNavList } from './SideNavList';
import { SideNavListItem } from './SideNavListItem';

// Icons and Logos
import { GoHomeFill } from "react-icons/go";
import { BiLibrary } from "react-icons/bi";
import { MdCategory } from "react-icons/md";
import { TbEdit } from "react-icons/tb";

export default function SideNavigation() {

  return (
    <SideNavWrapper className="" ariaLabel="Main navigation">
      <SideNavList>
        <SideNavListItem pageRoute="/library" icon={GoHomeFill}>Home</SideNavListItem>
        <SideNavListItem pageRoute="/library" icon={BiLibrary}>Library</SideNavListItem>
        <SideNavListItem pageRoute="/library/categories" icon={MdCategory}>Categories</SideNavListItem>
        <SideNavListItem pageRoute="/library/books/add" icon={TbEdit}>Add</SideNavListItem>
      </SideNavList>
    </SideNavWrapper>
  );
}
