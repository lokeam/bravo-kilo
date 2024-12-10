import { useCallback, useEffect, useState, useRef } from 'react';

interface MenuPosition {
  top: number;
  left: number;
}

interface UseMenuPositionProps {
  anchorEl: HTMLElement | null;
  menuId: string;
  align?: 'left' | 'right';
  offset?: { x?: number; y?: number };
  open: boolean;
}

export function useMenuPosition({
  anchorEl,
  menuId,
  align = 'left',
  offset = { x: 0, y: 0 },
  open
}: UseMenuPositionProps): MenuPosition {
  const [position, setPosition] = useState<MenuPosition>({ top: 0, left: 0 });
  const previousPosition = useRef<MenuPosition>({ top: 0, left: 0 });
  const updateCount = useRef(0);

  const updatePosition = useCallback(() => {
    if (!anchorEl) return;

    const rect = anchorEl.getBoundingClientRect();
    const menuElement = document.getElementById(menuId);
    if (!menuElement) return;

    const menuRect = menuElement.getBoundingClientRect();
    const viewportHeight = window.innerHeight;
    const viewportWidth = window.innerWidth;
    const scrollY = window.scrollY;
    const scrollX = window.scrollX;

    // Calculate initial position
    // Position directly under the anchor element
    let top = rect.bottom + 8; // Add 8px gap below icon
    let left = rect.left - (menuRect.width - rect.width) / 2; // Center align with icon

    // Ensure menu stays within viewport
    if (rect.bottom + menuRect.height > viewportHeight) {
      top = rect.top + scrollY - menuRect.height - (offset.y || 0);
    }

    if (left + menuRect.width > viewportWidth + scrollX) {
      left = viewportWidth + scrollX - menuRect.width - 8;
    }
    if (left < scrollX) {
      left = scrollX + 8;
    }

    // Only update if position has changed significantly (more than 1px)
    const hasSignificantChange =
      Math.abs(previousPosition.current.top - top) > 1 ||
      Math.abs(previousPosition.current.left - left) > 1;

    if (hasSignificantChange) {
      previousPosition.current = { top, left };
      setPosition({ top, left });

      if (process.env.NODE_ENV === 'development') {
        updateCount.current += 1;
        console.log(`Position updated ${updateCount.current} times`, { top, left });
      }
    }
  }, [anchorEl, menuId, align, offset]);

  // Reset counter when menu opens
  useEffect(() => {
    if (open) {
      updateCount.current = 0;
    }
  }, [open]);

  // Only update position when menu opens
  useEffect(() => {
    if (open) {
      // Initial position update
      updatePosition();

      // Resize handler with debounce
      let resizeTimeout: NodeJS.Timeout;
      const handleResize = () => {
        clearTimeout(resizeTimeout);
        resizeTimeout = setTimeout(updatePosition, 100);
      };

      window.addEventListener('resize', handleResize);

      return () => {
        window.removeEventListener('resize', handleResize);
        clearTimeout(resizeTimeout);
      };
    }
  }, [open, updatePosition]);

  return position;
}