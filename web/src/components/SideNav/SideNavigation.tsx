import SideNavWrapper from './SideNavWrapper';
import SideNavList from './SideNavList';
import SideNavListItem from './SideNavListItem';

// Icons and Logos
import { GoHomeFill } from "react-icons/go";
import { BiLibrary } from "react-icons/bi";
import { TbEdit } from "react-icons/tb";
import { IoSearchOutline } from 'react-icons/io5';

function SideNavigation() {
  return (
    <SideNavWrapper
      ariaLabel="Main navigation"
      className="side_nav_wrapper"
    >
      <SideNavList>
        <SideNavListItem
          icon={GoHomeFill}
          pageRoute="/home"
        >
          Home
        </SideNavListItem>
        <SideNavListItem
          icon={BiLibrary}
          pageRoute="/library"
        >
          Library
        </SideNavListItem>
        <SideNavListItem
          icon={IoSearchOutline}
          pageRoute="/library/books/search"
        >
          Search
        </SideNavListItem>
        <SideNavListItem
          icon={TbEdit}
          pageRoute="/library/books/add/gateway"
        >
          Add
        </SideNavListItem>
      </SideNavList>
    </SideNavWrapper>
  );
}

export default SideNavigation;
