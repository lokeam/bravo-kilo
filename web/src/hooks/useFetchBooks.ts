import useFetchData from './useFetchData';
import { fetchUserBooks } from '../service/apiClient.service';
import { Book } from '../types/api';

const useFetchBooks = (userID?: number, enabled: boolean = true) => {
  return useFetchData<Book[], number>('userBooks', fetchUserBooks, userID ?? 0, !!userID && enabled);
};

export default useFetchBooks;
