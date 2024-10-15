import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Book } from '../../types/api';

interface CardListItemGenreProps {
  genreName: string;
  genreImgs: string[];
  books: Book[];
}

export default function CardListItemGenre({ genreName, genreImgs, books }: CardListItemGenreProps) {

  const renderThumbnail = () => {
    if (genreImgs.length < 4) {
      return (
      <img
        alt={`Genre thumbnail for ${genreName}`}
        className="object-cover flex-none rounded w-16 h-16"
        loading="lazy"
        src={genreImgs[0]}
      />);
    } else {
      return (
        <div className="grid grid-rows-2 grid-cols-2 gap-0.5 w-16 h-16">
          {genreImgs.slice(0, 4).map((_, index) => (
            <img
              alt={`Genre thumbnail for ${genreName}`}
              className="h-auto w-full"
              key={`${index}-${genreName}`}
              loading="lazy"
              src={genreImgs[index]}
            />
          ))}
        </div>
      );
    }
  };

  const genreID = encodeURIComponent(genreName.split(' ').join('-'));
  const navigate = useNavigate();

  return (
      <motion.li
        className="py-3 flex items-start justify-between"
        key={genreName}
        layout
        onClick={() => navigate(`/library/${genreID}`, { state: books })}
      >
        <div className="flex gap-3 cursor-pointer">
          <div className="flex flex-row items-center justify-center rounded w-16 h-16 bg-dark-gunmetal">
            {renderThumbnail()}
          </div>
          <div className="card_list__item_copy flex flex-row items-center justify-center text-left pt-1">
            <span className="block text-base text-black dark:text-white font-semibold">{genreName}</span>
          </div>
        </div>
      </motion.li>
  );
}
