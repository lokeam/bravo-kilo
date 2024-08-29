import React, { useEffect, useState } from 'react';
import './Modal.css';

interface ModalProps {
  opened: boolean;
  onClose: () => void;
  children: React.ReactNode;
}

const Modal: React.FC<ModalProps> = ({ opened, onClose, children }) => {
  const [closing, setClosing] = useState(false);

  useEffect(() => {
    if (opened) {
      document.body.classList.add('modal-open');
    } else {
      document.body.classList.remove('modal-open');
    }

    return () => {
      document.body.classList.remove('modal-open');
    };
  }, [opened]);

  const handleClose = () => {
    console.log('handleClose fired')
    setClosing(true);
    setTimeout(() => {
      setClosing(false);
      onClose();
    }, 250); // match the duration of hideBottom animation
  };

  if (!opened && !closing) return null;

// active
// fixed bottom-0 left-0 right-0 z-40 w-full p-4 overflow-y-auto transition-transform bg-white dark:bg-gray-800 transform-none
  return (
    <div
      className="modal-overlay overflow-hidden fixed inset-0 z-30"
      data-closing={closing}
      onClick={handleClose}
    >
      <div
        className={`modal-content w-full lg:w-[400px] absolute lg:relative ${closing ? "" : "bottom-0"}`}
        onClick={(e) => e.stopPropagation()}
      >
        {children}
      </div>
    </div>
  );
};

export default Modal;
