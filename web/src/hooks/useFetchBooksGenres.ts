import useFetchData from './useFetchData';
import { fetchBooksGenres } from '../service/apiClient.service';

const useFetchBooksGenres = (userID: number, enabled: boolean) => {
  return useFetchData<Record<string, number>, number>(
    'booksGenres',
    (userID) => {
      if (userID === undefined) {
        return Promise.reject(new Error("useFetchBooksGeneres, UserID is undefined"));
      }
      return fetchBooksGenres(userID);
    },
    userID || 0,
    enabled
  );
};

export default useFetchBooksGenres;
