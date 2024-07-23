import React from 'react';
import ModalRoot from './ModalRoot';
import ModalHeader from './ModalHeader';
import ModalBody from './ModalBody';
import './Modal.css';

interface ModalProps {
  opened: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
}

const Modal: React.FC<ModalProps> = ({ opened, onClose, title, children }) => {
  return (
    <ModalRoot opened={opened} onClose={onClose}>
      <div className="modal-container">
        <ModalHeader title={title} onClose={onClose} />
        <ModalBody>{children}</ModalBody>
      </div>
    </ModalRoot>
  );
};

export default Modal;
