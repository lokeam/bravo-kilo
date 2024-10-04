import useFetchData from './useFetchData';
import { fetchBooksFormat } from '../service/apiClient.service';

const useFetchBooksFormat = (userID: number | undefined, enabled: boolean) => {
  return useFetchData<Record<string, number>, number>(
    'booksFormat',
    (userID) => {
      if (userID === undefined) {
        return Promise.reject(new Error("useFetchBooksFormat, UserID is undefined"));
      }
      return fetchBooksFormat(userID);
    },
    userID || 0,
    enabled);
};

export default useFetchBooksFormat;
