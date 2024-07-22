import { Book } from '../../pages/library';
import { BsThreeDotsVertical } from "react-icons/bs";
import { useNavigate } from 'react-router-dom';

interface CardListItemProps {
  book: Book;
}

export default function CardListItem({ book }: CardListItemProps) {
  const { authors, id, imageLinks, pageCount, title } = book;
  const navigate = useNavigate();

  const handleBookClick = () => {
    navigate(`/library/books/${id}`, { state: { book} });
  };

  return (
      <li key={`${id}-${title}-${pageCount}`} className="py-3 flex items-start justify-between">
        <div className="flex gap-3 cursor-pointer" onClick={handleBookClick}>
          <img loading="lazy" src={imageLinks[0]} alt={`Book cover thumbnail for ${title}`} className="flex-none rounded w-16 h-16" />
          <div className="card_list__item_copy text-left pt-1">
            <span className="block text-sm text-white font-semibold">{title}</span>
            <span className="block text-sm text-gray-400">by {authors[0]}</span>
          </div>
        </div>
        <button className="bg-transparent">
          <BsThreeDotsVertical color="white" />
        </button>
      </li>
  );
}
