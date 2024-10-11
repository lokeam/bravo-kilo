import useFetchData from './useFetchData';
import { fetchHomepageData } from '../service/apiClient.service';
import { HomepageStatistics } from '../types/api';

const useFetchHomepageData = (userID: number | undefined, enabled: boolean) => {
  return useFetchData<HomepageStatistics, number>({
    queryKey: ['booksHomepage', userID],
    fetchFunction: (userID) => {
      if (userID === undefined) {
        return Promise.reject(new Error("useFetchHomepageData, userID is undefined"));
      }
      return fetchHomepageData(userID);
    },
    query: userID,
    enabled: enabled && userID !== undefined,
  });
};

export default useFetchHomepageData;