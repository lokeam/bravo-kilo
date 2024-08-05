import useFetchData from './useFetchData';
import { fetchBooksAuthors } from '../service/apiClient.service';
import { Book } from '../pages/Library';

const useFetchBookAuthors = (userID: number, enabled: boolean) => {
  return useFetchData<Book[], number>('userBooksAuthors', fetchBooksAuthors, userID, enabled);
};

export default useFetchBookAuthors;
