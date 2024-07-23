import React from 'react';
import './Modal.css';

interface ModalBodyProps {
  children: React.ReactNode;
}

const ModalBody: React.FC<ModalBodyProps> = ({ children }) => {
  return <div className="modal-body text-black">{children}</div>;
};

export default ModalBody;
