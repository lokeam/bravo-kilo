
interface SideNavWrapperProps {
  ariaLabel?: string | undefined;
  className?: string | undefined;
  children?: React.ReactNode;
}

export const SideNavWrapper = ({ ariaLabel, children }: SideNavWrapperProps) => {
    {/*  <nav className={`fixed bottom-0 left-0 md:top-0 md:h-screen w-24 transition-transform -translate-x-full border-r bg-black border-yankees-blue md:translate-x-0 z-50 text-white`} aria-label={ariaLabel}> */}
  return (

    <nav className={`flex flex-row w-full fixed bottom-0 left-0 md:top-0 md:h-screen md:w-24 border-t md:border-r bg-black border-ebony-clay md:translate-x-0 z-50 text-white`} aria-label={ariaLabel}>
      <div className="flex flex-row w-full md:flex-col content-center justify-between md:justify-center overflow-y-auto px-5 md:px-2 h-full bg-black">
        {children}
      </div>
    </nav>
  );
};
