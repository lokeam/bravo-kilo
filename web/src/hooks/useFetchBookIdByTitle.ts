import useFetchData from './useFetchData';
import { fetchBookIDByTitle } from '../service/apiClient.service';
import { Book } from '../types/api';

const useFetchBookIdByTitle = (bookTitle: string, enabled: boolean = true) => {
  return useFetchData<Book, string>(
    'bookTitle',
    (bookTitle) => {
      if (bookTitle === undefined) {
        return Promise.reject(new Error("bookTitle is undefined"))
      }
      return fetchBookIDByTitle(bookTitle);
    },
    bookTitle || '',
    enabled
  );
};

export default useFetchBookIdByTitle;
