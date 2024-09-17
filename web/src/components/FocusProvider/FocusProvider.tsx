import { createContext, useContext, useRef, ReactNode } from 'react';

interface FocusContextType {
  searchFocusRef: React.RefObject<HTMLAnchorElement>;
  addManualRef: React.RefObject<HTMLAnchorElement>;
}

const FocusContext = createContext<FocusContextType | null>(null);

export function FocusProvider({ children }: {children: ReactNode }) {
  const searchFocusRef = useRef<HTMLAnchorElement>(null);
  const addManualRef = useRef<HTMLAnchorElement>(null);

  return (
    <FocusContext.Provider value={{ searchFocusRef, addManualRef }}>
      {children}
    </FocusContext.Provider>
  );
}

export const useFocusContext = () => {
  const context = useContext(FocusContext);
  if (!context) {
    throw new Error('useFocusContext must be used within a FocusProvider');
  }

  return context;
};
