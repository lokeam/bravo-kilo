import useFetchData from './useFetchData';
import { fetchBookByID } from '../service/apiClient.service';
import { Book } from '../types/api';

const useFetchBookById = (bookID: string, enabled: boolean = true) => {
  return useFetchData<Book, string>(
    'book',
    (bookID) => {
      if (bookID === undefined) {
        return Promise.reject(new Error("useFetchBookById, bookID is undefined"));
      }
      return fetchBookByID(bookID);
    },
    bookID,
    enabled);
};

export default useFetchBookById;
