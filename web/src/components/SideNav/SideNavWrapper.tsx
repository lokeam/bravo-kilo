
interface SideNavWrapperProps {
  ariaLabel?: string | undefined;
  className?: string | undefined;
  children?: React.ReactNode;
}

function SideNavWrapper({ ariaLabel, children }: SideNavWrapperProps) {
  return (
    <nav
      aria-label={ariaLabel}
      className={`flex flex-row w-full fixed bottom-0 left-0 mdTablet:top-0 mdTablet:h-screen mdTablet:w-20 border-t mdTablet:border-none bg-black border-ebony-clay mdTablet:translate-x-0 z-30 text-white`}
    >
      <div className="flex flex-row w-full mdTablet:flex-col content-center justify-between mdTablet:justify-center overflow-y-auto px-5 mdTablet:px-2 h-full bg-black">
        {children}
      </div>
    </nav>
  );
}

export default SideNavWrapper;
