import useFetchData from './useFetchData';
import { fetchBookIDByTitle } from '../service/apiClient.service';
import { Book } from '../types/api';

const useFetchBookIdByTitle = (bookTitle: string, enabled: boolean = true) => {
  return useFetchData<Book, string>('bookID', fetchBookIDByTitle, bookTitle, enabled);
};

export default useFetchBookIdByTitle;
