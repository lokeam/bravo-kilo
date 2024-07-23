import React from 'react';
import './Modal.css';

interface ModalHeaderProps {
  title: string;
  onClose: () => void;
}

const ModalHeader: React.FC<ModalHeaderProps> = ({ title, onClose }) => {
  return (
    <div className="modal-header">
      <h2>{title}</h2>
      <button className="modal-close-button" onClick={onClose}>
        x
      </button>
    </div>
  );
};

export default ModalHeader;
