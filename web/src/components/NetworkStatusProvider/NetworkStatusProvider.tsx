import { createContext, useState, useEffect, useContext, ReactNode } from 'react';

const NetworkStatusContext = createContext<boolean>(true);

export function NetworkStatusProvider({ children }: {children: ReactNode}) {
  const [isOnline, setIsOnline] = useState<boolean>(navigator.onLine);

  useEffect(() => {
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    }
  }, []);

  return (
    <NetworkStatusContext.Provider value={isOnline}>
      { children }
    </NetworkStatusContext.Provider>
  );
}

export const useNetworkStatus = () => {
  return useContext(NetworkStatusContext);
};
