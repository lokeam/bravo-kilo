import useFetchData from './useFetchData';
import { fetchBooksGenres } from '../service/apiClient.service';

const useFetchBooksGenres = (userID: number, enabled: boolean) => {
  return useFetchData<Record<string, number>, number>('booksGenres', fetchBooksGenres, userID, enabled);
};

export default useFetchBooksGenres;
