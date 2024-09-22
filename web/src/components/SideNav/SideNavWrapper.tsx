
interface SideNavWrapperProps {
  ariaLabel?: string | undefined;
  className?: string | undefined;
  children?: React.ReactNode;
}

function SideNavWrapper({ ariaLabel, children }: SideNavWrapperProps) {
  return (
    <nav
      aria-label={ariaLabel}
      className={`sideNavWrapper bg-white flex flex-row w-full fixed bottom-0 left-0 mdTablet:fixed mdTablet:top-0 mdTablet:w-24 border-t mdTablet:border-none border-ebony-clay mdTablet:translate-x-0 mdTablet:h-dvh z-30 dark:bg-black dark:text-white`}
    >
      <div className="flex flex-row w-full mdTablet:flex-col content-center justify-between mdTablet:justify-center overflow-y-auto px-5 mdTablet:px-2 h-full">
        {children}
      </div>
    </nav>
  );
}

export default SideNavWrapper;
