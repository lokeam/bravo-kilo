import { useCallback, useState } from 'react';
import Modal from '../Modal/Modal';
import useStore from '../../store/useStore';
import { PiArrowsDownUp } from 'react-icons/pi';

type CardListSortHeaderProps = {
  sortedBooksCount: number;
}

function CardListSortHeader ({ sortedBooksCount }: CardListSortHeaderProps) {
  const [isModalOpened, setIsModalOpened] = useState(false);
  const { sortCriteria, sortOrder, setSort } = useStore();

  // Handle sorting logic
  const handleSort = useCallback(
    (criteria: "title" | "publishDate" | "author" | "pageCount") => {
      const order = sortOrder === 'asc' ? 'desc' : 'asc';
      setSort(criteria, order);
      setIsModalOpened(false);
    },
    [sortOrder, setSort]
  );

  const openModal = () => setIsModalOpened(true);
  const closeModal = () => setIsModalOpened(false);

  const sortButtonTitle = {
    'title': 'Title: A to Z',
    'author': 'Author: A to Z',
    'publishDate': 'Release date: New to Old',
    'pageCount': 'Page count: Short to Long',
  };

  return (
    <div className="flex flex-row relative w-full max-w-7xl justify-between items-center text-left text-white border-b-2 border-solid border-zinc-700 pb-6 mb-2">
      <div className="mt-1">{sortedBooksCount} volumes</div>

      <div className="flex flex-row">
        <button className="flex flex-row justify-between bg-transparent border border-gray-600" onClick={openModal}>
          <PiArrowsDownUp size={22} className="pt-1 mr-2" color="white" />
          <span>{sortButtonTitle[sortCriteria]}</span>
        </button>
      </div>

      <Modal opened={isModalOpened} onClose={closeModal} title="">
        <button onClick={() => handleSort("publishDate")} className="flex flex-row bg-transparent mr-1">
          Release date: New to Old
        </button>
        <button onClick={() => handleSort("pageCount")} className="flex flex-row bg-transparent mr-1">
          Page count: Short to Long
        </button>
        <button onClick={() => handleSort("title")} className="flex flex-row bg-transparent mr-1">
          Title: A to Z
        </button>
        <button onClick={() => handleSort("author")} className="flex flex-row bg-transparent mr-1">
          Author: A to Z
        </button>
      </Modal>
    </div>
  );
}

export default CardListSortHeader;