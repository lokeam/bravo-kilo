import React, { useEffect } from 'react';
import { FaCheckCircle } from "react-icons/fa";
import { BsFillXCircleFill } from "react-icons/bs";
import { MdDeleteForever } from "react-icons/md";


interface SnackbarProps {
  message: string;
  open: boolean;
  duration?: number;
  variant: 'added' | 'updated' | 'removed' | 'error';
  onClose: () => void;
}

const Snackbar = ({ message, open, duration = 3000, onClose, variant }: SnackbarProps) => {
  useEffect(() => {
    if (open) {
      const timer = setTimeout(() => {
        onClose();
      }, duration);

      return () => clearTimeout(timer);
    }
  }, [open, duration, onClose]);

  if (!open) return null;


  const variantStyles = {
    added: 'bg-green-600',
    updated: 'bg-blue-600',
    removed: 'bg-slate-600',
    error: 'bg-red-600'
  };

  const variantIcons = {
    added: <FaCheckCircle className="mr-2" size={20} />,
    updated: <FaCheckCircle className="mr-2" size={20} />,
    removed: <MdDeleteForever className="mr-2" size={20} />,
    error: <BsFillXCircleFill className="mr-2" size={20} />
  };

  return (
    <div className={`${variantStyles[variant]} fixed bottom-36 left-1/2 transform -translate-x-1/2 text-white px-4 py-2 rounded shadow-lg flex items-center`}>
      { variantIcons[variant] }
      <span>{ message }</span>
    </div>
  )
};

export default Snackbar;