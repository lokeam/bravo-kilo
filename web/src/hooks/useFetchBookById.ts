import useFetchData from './useFetchData';
import { fetchBookByID } from '../service/apiClient.service';
import { Book } from '../types/api';

const useFetchBookById = (bookID: string, enabled: boolean = true) => {
  return useFetchData<Book, string>('book', fetchBookByID, bookID, enabled);
};

export default useFetchBookById;
