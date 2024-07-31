import axios from 'axios';

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
        await apiClient.post('/auth/token/refresh');
        return apiClient(originalRequest);
      } catch (refreshError) {
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

export const fetchUserBooks = async (userID: number) => {
  const { data } = await apiClient.get(`/api/v1/user/books?userID=${userID}`);
  return data.books || [];
};

export const fetchBooksCount = async (userID: number) => {
  const { data } = await apiClient.get(`/api/v1/user/books/count?userID=${userID}`);
  return data || {};
};

export const searchBookAPI = async (query: string) => {
  const { data } = await apiClient.get(`/api/v1/books/search`, {
    params: { query },
  });
  return data.items || [];
};

export default apiClient;
