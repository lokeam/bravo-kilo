import { useState } from 'react';
import { Book } from '../../pages/library';
import Modal from '../Modal/ModalRoot';
import { useNavigate } from 'react-router-dom';

import { AiFillFolderAdd } from "react-icons/ai";
import { BsThreeDotsVertical } from "react-icons/bs";
import { FaHeart } from "react-icons/fa";
import { MdMenuBook } from "react-icons/md";
import { TbEdit } from "react-icons/tb";


interface CardListItemProps {
  book: Book;
}

export default function CardListItem({ book }: CardListItemProps) {
  const [opened, setOpened] = useState<boolean>(false);

  const { authors, id, imageLinks, pageCount, title } = book;
  const navigate = useNavigate();

  const handleBookClick = () => {
    navigate(`/library/books/${id}`, { state: { book} });
  };

  const openModal = () => setOpened(true);
  const closeModal = () => setOpened(false);

  return (
      <li key={`${id}-${title}-${pageCount}`} className="py-3 flex items-start justify-between">
        <div className="flex gap-3 cursor-pointer" onClick={handleBookClick}>
          <img loading="lazy" src={imageLinks[0]} alt={`Book cover thumbnail for ${title}`} className="flex-none rounded w-16 h-16" />
          <div className="card_list__item_copy text-left pt-1">
            <span className="block text-sm text-white font-semibold">{title}</span>
            <span className="block text-sm text-gray-400">by {authors[0]}</span>
          </div>
        </div>
        <button onClick={openModal} className="bg-transparent">
          <BsThreeDotsVertical color="white" />
        </button>
        <Modal opened={opened} onClose={closeModal} title="">
          <button className="flex flex-row justify-items-start items-center bg-transparent w-full mr-1">
            <MdMenuBook className="mr-8" size={25}/>
            <span>Title Details</span>
          </button>
          <button className="flex flex-row justify-items-start items-center bg-transparent w-full mr-1">
            <TbEdit className="mr-8" size={25}/>
            <span>Edit Title Details</span>
          </button>
          <button className="flex flex-row justify-items-start items-center bg-transparent w-full mr-1">
            <FaHeart className="mr-9" size={20}/>
            <span>Add to Favorites</span>
          </button>
          <button className="flex flex-row justify-items-start items-center bg-transparent w-full mr-1">
            <AiFillFolderAdd className="mr-8" size={25}/>
            <span>Add to...</span>
          </button>
        </Modal>
      </li>
  );
}
