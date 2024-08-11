import { useState } from 'react';
import Modal from '../Modal/ModalRoot';
import ImagePlaceholder from './ImagePlaceholder';
import { useNavigate } from 'react-router-dom';
import { Book } from '../../pages/Library';

import { AiFillFolderAdd } from "react-icons/ai";
import { BsThreeDotsVertical } from "react-icons/bs";
import { FaHeart } from "react-icons/fa";
import { MdMenuBook } from "react-icons/md";
import { TbEdit } from "react-icons/tb";


interface CardListItemProps {
  book: Book;
  isSearchPage?: boolean;
}

export default function CardListItem({ book, isSearchPage }: CardListItemProps) {
  const [opened, setOpened] = useState<boolean>(false);

  const { authors, id, imageLinks, title } = book;
  const titleSubdomain = encodeURIComponent(title);
  const navigate = useNavigate();

  //console.log('encoded titleParam: ', titleParam);


  const handleBookClick = () => {
    navigate(`/library/books/${titleSubdomain}`, { state: { book } });
  };

  const openModal = () => setOpened(true);
  const closeModal = () => setOpened(false);

  return (
      <li key={`${title}-${id}`} className="py-3 flex items-start justify-between">
        <div className="flex gap-3 cursor-pointer" onClick={handleBookClick}>
          {
            imageLinks.length > 0 ? (
              <img loading="lazy" src={imageLinks[0]} alt={`Book cover thumbnail for ${title}`} className="flex-none rounded w-16 h-16" />
            ) : (
              <ImagePlaceholder isBookDetail={false}/>
            )
          }
          <div className="card_list__item_copy text-left pt-1">
            <span className="block text-base text-white font-semibold">{title}</span>
            <span className="block text-sm text-gray-400">by {authors[0]}</span>
            {
              isSearchPage && book.isInLibrary && <div className="block text-sm text-white font-semibold">IN YOUR LIBRARY</div>
            }
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
