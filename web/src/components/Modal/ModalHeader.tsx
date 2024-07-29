import React from 'react';
import './Modal.css';

interface ModalHeaderProps {
  title?: string;
  onClose: () => void;
}

const ModalHeader: React.FC<ModalHeaderProps> = ({ title, onClose }) => {
  return (
    <div className="modal-header border-nevada-gray border-t">
      <h2>{title}</h2>
      <button className="modal-close-button w-full bg-transparent active:border-transparent focus:border-transparent hover:border-transparent md:border-transparent" onClick={onClose}>
        Close
      </button>
    </div>
  );
};

export default ModalHeader;
