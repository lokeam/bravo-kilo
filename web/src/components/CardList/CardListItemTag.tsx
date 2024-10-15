import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Book } from '../../types/api';

interface CardListItemTagProps {
  tagName: string;
  tagImgs: string[];
  books: Book[];
}

export default function CardListItemTag({ tagName, tagImgs, books }: CardListItemTagProps) {

  const renderThumbnail = () => {
    if (tagImgs.length < 4) {
      return (
      <img
        alt={`Genre thumbnail for ${tagName}`}
        className="object-cover flex-none rounded w-16 h-16"
        loading="lazy"
        src={tagImgs[0]}
      />);
    } else {
      return (
        <div className="grid grid-rows-2 grid-cols-2 gap-0.5 w-16 h-16">
          {tagImgs.slice(0, 4).map((_, index) => (
            <img
              alt={`Genre thumbnail for ${tagName}`}
              className="h-auto w-full"
              key={`${index}-${tagName}`}
              loading="lazy"
              src={tagImgs[index]}
            />
          ))}
        </div>
      );
    }
  };

  const genreID = encodeURIComponent(tagName.split(' ').join('-'));
  const navigate = useNavigate();

  return (
      <motion.li
        className="py-3 flex items-start justify-between"
        key={tagName}
        layout
        onClick={() => navigate(`/library/${genreID}`, { state: books })}
      >
        <div className="flex gap-3 cursor-pointer">
          <div className="flex flex-row items-center justify-center rounded w-16 h-16 bg-dark-gunmetal">
            {renderThumbnail()}
          </div>
          <div className="card_list__item_copy flex flex-row items-center justify-center text-left pt-1">
            <span className="block text-base text-black dark:text-white font-semibold">{tagName}</span>
          </div>
        </div>
      </motion.li>
  );
}
