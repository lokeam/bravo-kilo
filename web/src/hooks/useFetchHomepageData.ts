import useFetchData from './useFetchData';
import { fetchHomepageData } from '../service/apiClient.service';

const useFetchHomepageData = (userID: number | undefined, enabled: boolean) => {
  return useFetchData('booksHomepage', fetchHomepageData, userID, enabled);
};

export default useFetchHomepageData;
