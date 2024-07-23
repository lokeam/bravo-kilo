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
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'auto';
    }
    return () => {
      document.body.style.overflow = 'auto';
    };
  }, [opened]);

  const handleClose = () => {
    setClosing(true);
    setTimeout(() => {
      setClosing(false);
      onClose();
    }, 250); // match the duration of hideBottom animation
  };

  if (!opened && !closing) return null;

  return (
    <div className="modal-overlay" data-closing={closing} onClick={handleClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        {children}
      </div>
    </div>
  );
};

export default Modal;
