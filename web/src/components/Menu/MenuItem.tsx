import { memo } from 'react';

interface MenuItemProps {
  label: string;
  onClick: () => void;
  isHighlighted: boolean;
  onMouseEnter: (index: number) => void;
  index: number;
  isActive?: boolean;
}

function MenuItemComponent({
  label,
  onClick,
  isHighlighted,
  onMouseEnter,
  index,
  isActive,
}: MenuItemProps) {

  return (
    <li
      className={`
        px-4 py-2 cursor-pointer
        ${isHighlighted ? 'bg-gray-100 dark:bg-gray-700' : 'bg-transparent'}
        transition-colors duration-150 ease-in-out
        hover:bg-gray-100 dark:hover:bg-gray-700
        ${isActive ? 'bg-gray-200 dark:bg-gray-600 text-vivid-blue font-semibold' : ''}
        `}
      role="menuitem"
      onClick={onClick}
      onMouseEnter={() => onMouseEnter(index)}
      data-index={index}
    >
      {label}
    </li>
  );
}

// Memoize MenuItem with strict comparison
export const MenuItem = memo(MenuItemComponent, (prevProps, nextProps) => {
  return (
    prevProps.label === nextProps.label &&
    prevProps.isHighlighted === nextProps.isHighlighted &&
    prevProps.index === nextProps.index &&
    prevProps.onClick === nextProps.onClick &&
    prevProps.onMouseEnter === nextProps.onMouseEnter &&
    prevProps.isActive === nextProps.isActive
  );
});