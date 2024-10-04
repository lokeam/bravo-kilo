import useFetchData from './useFetchData';
import { fetchHomepageData } from '../service/apiClient.service';

const useFetchHomepageData = (userID: number | undefined, enabled: boolean) => {
  return useFetchData(
    'booksHomepage',
    (userID) => {
      if (userID === undefined) {
        return Promise.reject(new Error("useFetchHomepageData, userID is undefined"));
      }
      return fetchHomepageData(userID);
    },
    userID,
    enabled);
};

export default useFetchHomepageData;
