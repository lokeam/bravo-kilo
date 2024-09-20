import React from 'react';
import { useNetworkStatus } from '../NetworkStatusProvider/NetworkStatusProvider';
import { MdOutlineSignalWifiConnectedNoInternet4 } from "react-icons/md";


const OfflineBanner = () => {
  const isOnline = useNetworkStatus();

  return (
    !isOnline && (
      <div className="offline_banner bg-charcoal relative box-border flex flex-col min-h-min items-center content-center h-full py-4 px-2">
        <div className="flex flex-col w-full items-center content-center">
          <div className="w-64 flex flex-row">
            <div className="relative inline-flex items-center justify-center overflow-hidden bg-transparent h-9 w-9 rounded-full">
              <MdOutlineSignalWifiConnectedNoInternet4 color="#fff" size={30} />
          </div>
            <div className="flex flex-col items-center justify-center text-lg font-bold text-center align-middle ml-3">No Internet Connection</div>
          </div>
        </div>
      </div>
    )
  );
};

export default OfflineBanner;
