
interface SideNavWrapperProps {
  ariaLabel?: string | undefined;
  className?: string | undefined;
  children?: React.ReactNode;
}

export const SideNavWrapper = ({ ariaLabel, children }: SideNavWrapperProps) => {
  return (
    <nav className={`flex flex-row w-full fixed bottom-0 left-0 md:top-0 md:h-screen md:w-20 border-t md:border-none bg-black border-ebony-clay md:translate-x-0 z-30 text-white`} aria-label={ariaLabel}>
      <div className="flex flex-row w-full md:flex-col content-center justify-between md:justify-center overflow-y-auto px-5 md:px-2 h-full bg-black">
        {children}
      </div>
    </nav>
  );
};
