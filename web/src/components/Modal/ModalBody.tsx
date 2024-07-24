import React from 'react';
import './Modal.css';

interface ModalBodyProps {
  children: React.ReactNode;
}

const ModalBody: React.FC<ModalBodyProps> = ({ children }) => {
  return <div className="modal-body flex flex-col p-5">{children}</div>;
};

export default ModalBody;
