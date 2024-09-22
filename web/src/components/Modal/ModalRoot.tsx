import React, { useEffect, useState } from 'react';
import './Modal.css';

interface ModalProps {
  opened: boolean;
  onClose: () => void;
  children: React.ReactNode;
}

function Modal({ opened, onClose, children }: ModalProps) {
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

  return (
    <div
      className="modal-overlay overflow-hidden top-0 left-0 inset-0 z-30 opacity-100"
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
}

export default Modal;
