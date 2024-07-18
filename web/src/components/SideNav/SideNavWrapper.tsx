
interface SideNavWrapperProps {
  ariaLabel?: string | undefined;
  className?: string | undefined;
  children?: React.ReactNode;
}

export const SideNavWrapper = ({ ariaLabel, children }: SideNavWrapperProps) => {
  return (
    <aside className={`fixed top-0 left-0 w-64 h-screen transition-transform -translate-x-full border-r bg-yankees-blue border-yankees-blue md:translate-x-0 z-50 p-5 text-white`} aria-label={ariaLabel}>
      <div className="overflow-y-auto px-3 h-full bg-yankees-blue">
        {children}
      </div>
    </aside>
  );
};
