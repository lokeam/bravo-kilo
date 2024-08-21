import useFetchData from './useFetchData';
import { fetchHomepageData } from '../service/apiClient.service';

const useFetchHomepageData = (userID: number, enabled: boolean) => {
  return useFetchData<Record<string, number>, number>('booksHomepage', fetchHomepageData, userID, enabled);
};

export default useFetchHomepageData;
