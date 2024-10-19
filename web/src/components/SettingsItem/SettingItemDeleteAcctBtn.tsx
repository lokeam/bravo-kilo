import { useState, useEffect } from "react";
import { IoIosWarning } from "react-icons/io";
import Modal from "../Modal/Modal";
import { MdDeleteForever } from "react-icons/md";

function SettingsItemDeleteAcctBtn() {
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);
  const [confirmText, setConfirmText] = useState<string>('');
  const [isDeleteEnabled, setDeleteEnabled] = useState<boolean>(false);

  const openModal = () => setIsModalOpen(true);
  const closeModal = () => {
    setIsModalOpen(false);
    setConfirmText('');
    setDeleteEnabled(false);
  };

  const handleDelete = () => {
    console.log('Account deletion activated');
    closeModal();
  };

  useEffect(() => {
    setDeleteEnabled(confirmText.toLowerCase() === 'delete my account');
  }, [confirmText]);

  return (
    <>
      <button
        onClick={openModal}
        className="bg-gray-200 h-11 border-red-500 text-red-500 hover:text-white hover:bg-red-600 hover:border-red-600 focus:ring-red-900 dark:bg-gray-800 dark:hover:bg-red-600 dark:border-2 transition duration-500 ease-in-out"
      >
        Delete my account
      </button>
      <Modal
        opened={isModalOpen}
        onClose={closeModal}
      >

        <div className="flex items-center justify-center">
          <IoIosWarning className="text-orange-600 dark:text-yellow-500" size={30} />
        </div>
        <h2 className="text-2xl text-center font-semibold text-red-800 pb-2">Danger Zone</h2>
        <h3 className="flex items-center justify-center text-lg">Are you sure that you want to delete your account?</h3>
        <p className="font-bold flex items-center justify-center mb-5">This action cannot be undone.</p>
        <p className="text-s  m text-gray-600 dark:text-gray-400 mb-2">To confirm, type <span className="font-bold text-red-800 dark:text-red-500">delete my account</span> below.</p>
        <input
          type="text"
          value={confirmText}
          onChange={(e) => setConfirmText(e.target.value)}
          className="border-gray-700/60 bg-transparent border-2 rounded-md p-2 mb-2"
          placeholder="Type 'delete my account'"
        />
        <button
          type="button"
          onClick={closeModal}
          className="flex flex-row justify-between items-center bg-transparent border-gray-700/60 mr-1 w-full mb-3 lg:mb-0 transition duration-500 ease-in-out hover:border-vivid-blue dark:hover:border-vivid-blue hover:bg-vivid-blue dark:hover:bg-vivid-blue hover:text-white dark:hover:text-whtie"
        >
          <span>Cancel</span>
        </button>
        <button
          type="button"
          onClick={handleDelete}
          disabled={!isDeleteEnabled}
          className={`bg-transparent flex flex-row justify-between items-center mr-1 w-full border-red-500 text-red-500 hover:text-white dark:hover:text-white hover:bg-red-800 focus:ring-red-800 hover:border-red-800 dark:hover:bg-red-800 dark:hover:border-red-800 transition duration-500 ease-in-out disabled:pointer-events-none disabled:border-gray-700/60 disabled:text-gray-600 dark:disabled:text-gray-400`}
        >
          <span>Yes, I want to delete my account</span>
          <MdDeleteForever size={30}/>
        </button>
      </Modal>
    </>
  );
}

export default SettingsItemDeleteAcctBtn;
