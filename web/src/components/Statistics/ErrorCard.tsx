import { IoCloudOfflineOutline } from "react-icons/io5";

const ErrorCard = () => {
  return (
    <div className="flex flex-col w-full items-center content-center justify-center h-full">
      <IoCloudOfflineOutline
        className="mb-6"
        size={50}
      />
      <div className="flex flex-col w-full items-center content-center">
        <h2 className="text-2xl pb-6">Connect to the Internet</h2>
        <p className="pb-6">Your current network connection is offline.</p>
        <p>Please check your connection.</p>
      </div>
    </div>
  )
}

export default  ErrorCard;
