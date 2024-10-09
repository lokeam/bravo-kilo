import React from 'react';
import ModalRoot from './ModalRoot';
import ModalBody from './ModalBody';
import './Modal.css';

interface ModalProps {
  opened: boolean;
  onClose: () => void;
  title?: string;
  children: React.ReactNode;
}

function Modal({ opened, onClose, children }: ModalProps) {
  return (
    <ModalRoot
      opened={opened}
      onClose={onClose}
    >
      <div className="modal-container">
        <ModalBody>{children}</ModalBody>
      </div>
    </ModalRoot>
  );
}

export default Modal;
