import useFetchData from './useFetchData';
import { fetchUserBooks } from '../service/apiClient.service';
import { Book } from '../types/api';

const useFetchBooks = (userID: number | undefined, enabled: boolean = true) => {
  return useFetchData<Book[], number>({
    queryKey: ['userBooks', userID],
    fetchFunction: (userID) => {
      if (userID === undefined) {
        return Promise.reject(new Error("useFetchBooks, UserID is undefined"));
      }
      return fetchUserBooks(userID);
    },
    query: userID,
    enabled: enabled && userID !== undefined,
  });
};

export default useFetchBooks;