import useFetchData from './useFetchData';
import { fetchBooksFormat } from '../service/apiClient.service';
import { BookFormatData } from '../types/api';

const useFetchBooksFormat = (userID: number | undefined, enabled: boolean) => {
  return useFetchData<BookFormatData, number>({
    queryKey: ['booksFormat', userID],
    fetchFunction: (userID) => {
      if (userID === undefined) {
        return Promise.reject(new Error("useFetchBooksFormat, UserID is undefined"));
      }
      return fetchBooksFormat(userID);
    },
    query: userID,
    enabled: enabled && userID !== undefined,
  });
};

export default useFetchBooksFormat;