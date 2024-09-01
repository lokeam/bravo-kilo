import useFetchData from './useFetchData';
import { fetchBooksFormat } from '../service/apiClient.service';

const useFetchBooksFormat = (userID: number | undefined, enabled: boolean) => {
  return useFetchData('booksFormat', fetchBooksFormat, userID, enabled);
};

export default useFetchBooksFormat;
