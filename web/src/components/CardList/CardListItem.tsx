import { useState } from 'react';
import Modal from '../Modal/ModalRoot';
import ImagePlaceholder from './ImagePlaceholder';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion'
import { Book } from '../../types/api';

import { BsThreeDotsVertical } from "react-icons/bs";
import { MdMenuBook } from "react-icons/md";
import { TbEdit } from "react-icons/tb";


interface CardListItemProps {
  book: Book;
  isSearchPage?: boolean;
}

const isWhitelistedImageURL = (imageURL: string): boolean => {
  const allowedDomains = ['google.com', 'unsplash.com'];
  try {
    const url = new URL(imageURL);
    return allowedDomains.some(domain => url.hostname.endsWith(domain));
  } catch {
    return false;
  }
}

function CardListItem({ book, isSearchPage }: CardListItemProps) {
  const [opened, setOpened] = useState<boolean>(false);
  const navigate = useNavigate();
  const { authors, id, imageLink, title } = book;
  const titleSubdomain = encodeURIComponent(title);

  //console.log('card list item, book: ', book);

  const handleBookClick = () => {
    navigate(`/library/books/${titleSubdomain}`, { state: { book } });
  };

  const handleEditBookClick = () => {
    navigate(`/library/books/${id}/edit`, { state: { book } });
  };

  const openModal = () => setOpened(true);
  const closeModal = () => setOpened(false);

  const hasImageLink = imageLink && imageLink !== "";
  const mayRenderImage = hasImageLink && isWhitelistedImageURL(imageLink);

  return (
      <motion.li
        className="py-3 flex items-start justify-between"
        key={`${title}-${id}`}
        layout
      >
        <div
          className="flex object-cover gap-3 cursor-pointer"
          onClick={handleBookClick}
        >
          {
            mayRenderImage ? (
              <img
                alt={`Book cover thumbnail for ${title}`}
                className="object-cover flex-none rounded w-16 h-16"
                loading="lazy"
                src={imageLink}
              />
            ) : (
              <ImagePlaceholder isBookDetail={false}/>
            )
          }
          <div className="card_list__item_copy text-left pt-1">
            <span className="block text-base text-black dark:text-white font-semibold">{title}</span>
            <span className="block text-sm text-gray-500 dark:text-gray-400">by {authors && authors.length > 0 ? authors[0] : 'author data not available'}</span>
            {
              isSearchPage && book.isInLibrary && <div className="block text-sm text-white font-semibold">IN YOUR LIBRARY</div>
            }
          </div>
        </div>
        <button
          onClick={openModal}
          className="bg-transparent border-charcoal dark:border-2 transition duration-500 ease-in-out dark:hover:border-vivid-blue"
        >
          <BsThreeDotsVertical
            className="text-charcoal dark:text-white"
          />
        </button>
        <Modal
          opened={opened}
          onClose={closeModal}
          title=""
        >
          <button
            className="flex flex-row justify-items-start items-center bg-transparent w-full mr-1"
            onClick={handleBookClick}
          >
            <MdMenuBook
              className="mr-8"
              size={25}
            />
            <span>Title Details</span>
          </button>
          <button
            className="flex flex-row justify-items-start items-center bg-transparent w-full mr-1"
            onClick={handleEditBookClick}
            >
            <TbEdit
              className="mr-8 dark:text-white"
              size={25}
            />
            <span>Edit Title Details</span>
          </button>
        </Modal>
      </motion.li>
  );
}

export default CardListItem;
