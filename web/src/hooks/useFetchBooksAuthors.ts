import useFetchData from './useFetchData';
import { fetchBooksAuthors } from '../service/apiClient.service';
import { Book } from '../types/api';

const useFetchBookAuthors = (userID: number | undefined, enabled: boolean) => {
  return useFetchData<Book[], number>(
    'userBooksAuthors',
    (userID) => {
      if (userID === undefined) {
        return Promise.reject(new Error("useFetchBooksAuthors, UserID is undefined"));
      }
      return fetchBooksAuthors(userID);
    },
    userID,
    enabled);
};

export default useFetchBookAuthors;
