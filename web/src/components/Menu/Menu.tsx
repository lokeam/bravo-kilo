import { useCallback, useEffect, useState, memo, useMemo } from 'react';
import { MenuItem } from './MenuItem';
import { useMenuPosition } from './useMenuPosition';

interface MenuProps {
  open: boolean;
  anchorEl: HTMLElement | null;
  onClose: () => void;
  items: {
    label: string;
    onClick: () => void;
  }[];
  align?: 'left' | 'right';
  offset?: { x?: number; y?: number };
  isActive?: boolean;
}

interface MenuListProps {
  items: MenuProps['items'];
  onItemClick: (index: number) => void;
}

// Define MenuList component
function MenuListComponent({
  items,
  onItemClick,
}: MenuListProps) {
  const [highlightedIndex, setHighlightedIndex] = useState<number>(0);

  const handleMouseEnter = useCallback((index: number) => {
    setHighlightedIndex(index);
  }, []);

  const itemCallbacks = useMemo(() =>
    items.map((_, index) => ({
      onClick: () => onItemClick(index),
      onMouseEnter: () => handleMouseEnter(index)
    })), [items, onItemClick, handleMouseEnter]);

  return (
    <>
      {items.map((item, index) => (
        <MenuItem
          key={`${item.label}-${index}`}
          label={item.label}
          onClick={itemCallbacks[index].onClick}
          isHighlighted={index === highlightedIndex}
          onMouseEnter={itemCallbacks[index].onMouseEnter}
          index={index}
          isActive={item.isActive}
        />
      ))}
    </>
  );
}

// Memoize MenuList
const MenuList = memo(MenuListComponent);

// Define Menu component
function MenuComponent({
  open,
  anchorEl,
  onClose,
  items,
  align = 'left',
  offset = { x: 0, y: 0 },
}: MenuProps) {
  const position = useMenuPosition({
    anchorEl,
    menuId: 'menu-list',
    align,
    offset,
    open
  });

  const handleClick = useCallback((index: number) => {
    items[index].onClick();
    onClose();
  }, [items, onClose]);

  // Click outside handler
  const handleClickOutside = useCallback((event: MouseEvent) => {
    if (
      anchorEl &&
      !anchorEl.contains(event.target as Node) &&
      event.target instanceof Node &&
      !document.getElementById('menu-list')?.contains(event.target)
    ) {
      onClose();
    }
  }, [anchorEl, onClose]);

  // Add/remove click outside listener
  useEffect(() => {
    if (open) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [open, handleClickOutside]);

  if (!open || !anchorEl) return null;

  return (
    <ul
      id="menu-list"
      aria-labelledby="menu-button"
      className="fixed bg-white dark:bg-gray-800 shadow-lg rounded-md py-1 z-50 min-w-[200px]"
      role="menu"
      style={{
        top: `${position.top}px`,
        left: `${position.left}px`,
      }}
      tabIndex={-1}
    >
      <MenuList
        items={items}
        onItemClick={handleClick}
      />
    </ul>
  );
}

// Memoize Menu with comparison function
export const Menu = memo(MenuComponent, (prevProps, nextProps) => {
  return (
    prevProps.open === nextProps.open &&
    prevProps.anchorEl === nextProps.anchorEl &&
    prevProps.items === nextProps.items &&
    prevProps.align === nextProps.align &&
    prevProps.offset?.x === nextProps.offset?.x &&
    prevProps.offset?.y === nextProps.offset?.y
  );
});