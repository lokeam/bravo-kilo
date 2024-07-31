import axios from 'axios';
import { Book } from '../pages/Library';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_ENDPOINT,
  withCredentials: true,
});

apiClient.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config;

    if (error.response.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        console.log('attempting to refresh token');
        await apiClient.post('/auth/token/refresh');
        return apiClient(originalRequest);
      } catch (refreshError) {
        console.error('Token refresh failed', refreshError);
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

export default apiClient;

export const fetchUserBooks = async (userID: number): Promise<Book[]> => {
  const { data } = await apiClient.get(`/api/v1/user/books?userID=${userID}`);
  return data.books;
};

export const fetchBooksCount = async (userID: number) => {
  const { data } = await apiClient.get(`/api/v1/user/books/count?userID=${userID}`);
  return data;
};
