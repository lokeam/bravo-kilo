import useFetchData from './useFetchData';
import { fetchBooksCount } from '../service/apiClient.service';

const useFetchBooksCount = (userID: number, enabled: boolean) => {
  return useFetchData<Record<string, number>, number>('booksCount', fetchBooksCount, userID, enabled);
};

export default useFetchBooksCount;
