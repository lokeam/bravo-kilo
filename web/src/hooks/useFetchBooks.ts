import useFetchData from './useFetchData';
import { fetchUserBooks } from '../service/apiClient.service';
import { Book } from '../pages/Library';

const useFetchBooks = (userID: number) => {
  return useFetchData<Book[]>('userBooks', fetchUserBooks, userID);
};

export default useFetchBooks;
