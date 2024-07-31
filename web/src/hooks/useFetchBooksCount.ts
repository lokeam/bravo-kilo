import useFetchData from './useFetchData';
import { fetchBooksCount } from '../service/apiClient.service';

const useFetchBooksCount = (userID: number) => {
  return useFetchData('booksCount', fetchBooksCount, userID);
};

export default useFetchBooksCount;
