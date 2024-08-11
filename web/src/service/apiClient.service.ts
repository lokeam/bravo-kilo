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

      const refreshToken = document.cookie.split('; ').find(row => row.startsWith('refresh_token='));
      if (!refreshToken) {
        console.log('No refresh token present, redirecting to login.');
        window.location.href = '/login';
        return Promise.reject(error);
      }

      try {
        console.log('apiClient, trying token refresh');
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

export const fetchUserBooks = async (userID: number) => {
  const { data } = await apiClient.get(`/api/v1/user/books?userID=${userID}`);
  return data.books || [];
};

export const fetchBooksFormat = async (userID: number) => {
  const { data } = await apiClient.get(`/api/v1/user/books/format?userID=${userID}`);
  return data || {};
};

export const fetchBooksAuthors = async (userID: number) => {
  const { data } = await apiClient.get(`/api/v1/user/books/authors?userID=${userID}`);
  return data || {};
};

export const fetchBooksGenres = async (userID: number) => {
  const { data } = await apiClient.get(`/api/v1/user/books/genres?userID=${userID}`);
  return data || {};
}

export const searchBookAPI = async (query: string) => {
  const { data } = await apiClient.get(`/api/v1/books/search`, {
    params: { query },
  });
  console.log('searchBookAPI:', data); // Log the response
  return data || [];
};

export const fetchBookByID = async (bookID: string) => {
  const { data } = await apiClient.get(`/api/v1/books/by-id/${bookID}`);
  return data.book;
};

export const fetchBookIDByTitle = async(bookTitle: string) => {
  const { data } = await apiClient.get(`/api/v1/books/by-title`, {
    params: { title: bookTitle }
  });
  return data.bookID;
}

export const verifyUserToken = async () => {
  const { data } = await apiClient.get('/auth/token/verify');
  return data.user;
};

export const signOutUser = async () => {
  await apiClient.post('/auth/signout');
};

export const updateBook = async (book: Book, bookID: string) => {
  const { data } = await apiClient.put(`/api/v1/books/${bookID}`, book);
  return data;
};

export const addBook = async (book: Book) => {
  console.log('apiClient.service, about to post book from addBook:', book);
  const { data } = await apiClient.post('/api/v1/books/add', book);
  console.log('apiClient.service, received response from addBook:', data);
  return data;
}

export default apiClient;
