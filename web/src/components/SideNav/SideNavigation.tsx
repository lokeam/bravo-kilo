import { SideNavWrapper } from './SideNavWrapper';
import { SideNavLogo } from './SideNavLogo';
import { SideNavList } from './SideNavList';
import { SideNavListItem } from './SideNavListItem';
import { SideNavCollapse } from './SideNavCollapse';

// Icons and Logos
import { MdOutlineGridView } from 'react-icons/md';
import { HiOutlineCollection } from 'react-icons/hi';
import bkLogo from '../../assets/tk_icon.webp'

// interface SideNavState {
//   collapses: Record<string, boolean>;
// }

// interface SideNavProps {
//   sidenavState: SideNavState;
// }

export default function SideNavigation() {

  return (
    <SideNavWrapper className="" ariaLabel="Main navigation">
      <SideNavLogo href="#" img={bkLogo} imgAlt="KB logo" />

      <SideNavList>
        <SideNavCollapse
          label="View As"
          icon={MdOutlineGridView}
          collapseKey='viewAs'
        >
          <SideNavListItem href="#">Grid</SideNavListItem>
          <SideNavListItem href="#">Table</SideNavListItem>
          <SideNavListItem href="#">List</SideNavListItem>
        </SideNavCollapse>

        <SideNavCollapse
          label="Categories"
          icon={HiOutlineCollection}
          collapseKey='categories'
        >
          <SideNavListItem href="#">My Category #1</SideNavListItem>
          <SideNavListItem href="#">My Category #2</SideNavListItem>
          <SideNavListItem href="#">Starred</SideNavListItem>
          <SideNavListItem href="#">To Read</SideNavListItem>
        </SideNavCollapse>

      </SideNavList>
    </SideNavWrapper>
  );
}
