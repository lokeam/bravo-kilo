import { useNavigate } from "react-router-dom";
import { motion } from 'framer-motion';
import { Book } from "../../types/api";
import { BsPerson } from "react-icons/bs";

interface CardListItemAuthorProps {
  authorName: string;
  books: Book[];
}

export default function CardListItemAuthor({ authorName, books }: CardListItemAuthorProps) {

  const authorID = encodeURIComponent(authorName.split(' ').join('-'));

  console.log('----');
  console.log('testing joined authorID: ', authorID);
  console.log('testing author books: ', books);

  const navigate = useNavigate();

  return (
      <motion.li
        className="py-3 flex items-start justify-between"
        key={authorName}
        layout
        onClick={() => navigate(`/library/${authorID}?page=author`, { state: books })}
      >
        <div className="flex gap-3 cursor-pointer">
          <div className="flex flex-row items-center justify-center rounded-full w-16 h-16 bg-dark-gunmetal">
            <BsPerson size={40} />
          </div>
          <div className="card_list__item_copy flex flex-row items-center justify-center text-left pt-1">
            <span className="block text-base text-black dark:text-white font-semibold">{authorName}</span>
          </div>
        </div>
      </motion.li>
  );
}
