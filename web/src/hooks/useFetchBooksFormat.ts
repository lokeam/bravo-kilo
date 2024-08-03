import useFetchData from './useFetchData';
import { fetchBooksFormat } from '../service/apiClient.service';

const useFetchBooksFormat = (userID: number, enabled: boolean) => {
  return useFetchData<Record<string, number>, number>('booksFormat', fetchBooksFormat, userID, enabled);
};

export default useFetchBooksFormat;
