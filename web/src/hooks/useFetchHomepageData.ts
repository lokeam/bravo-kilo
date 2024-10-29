import useFetchData from './useFetchData';
import { fetchHomepageData } from '../service/apiClient.service';
import { HomepageStatistics } from '../types/api';

const useFetchHomepageData = (userID: number | undefined, enabled: boolean) => {
  return useFetchData<HomepageStatistics, number>({
    queryKey: ['booksHomepage', userID],
    fetchFunction: async (userID) => {
      if (userID === undefined) {
        return Promise.reject(new Error("useFetchHomepageData, userID is undefined"));
      }
      const data = await fetchHomepageData(userID);
      console.log('Homepage Data Response:', data);
      return data;
    },
    query: userID,
    enabled: enabled && userID !== undefined,
  });
};

export default useFetchHomepageData;