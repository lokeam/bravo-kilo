import useFetchData from './useFetchData';
import { fetchUserBooks } from '../service/apiClient.service';
import { Book } from '../pages/Library';

const useFetchBooks = (userID: number, enabled: boolean = true) => {
  return useFetchData<Book[], number>('userBooks', fetchUserBooks, userID, enabled);
};

export default useFetchBooks;
