import axios from 'axios';
import { Book } from '../types/api';

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
};

export const searchBookAPI = async (query: string) => {
  const { data } = await apiClient.get('/api/v1/books/search', {
    params: { query },
  });
  console.log('searchBookAPI:', data); // Log the response
  return data || [];
};

export const geminiQueryAPI = async (prompt: string) => {
  console.log('apiClient.service - geminiQueryAPI get req fired');
  const { data } = await apiClient.get('/api/v1/books/summary', {
    params: { prompt },
  })
  console.log('geminiQueryAPI response: ', data);
  return data || {};
};

export const fetchBookByID = async (bookID: string) => {
  const { data } = await apiClient.get(`/api/v1/books/by-id/${bookID}`);
  return data.book;
};

export const fetchBookIDByTitle = async (bookTitle: string) => {
  try {
    const response = await apiClient.get('/api/v1/books/by-title', {
      params: { title: bookTitle }
    });
    //console.log('API response:', response.data);
    return response.data.bookID; // Ensure the structure matches
  } catch (error) {
    console.error('Error fetching book ID:', error);
    throw new Error('Failed to fetch book ID');
  }
};

export const fetchHomepageData = async (userID: number) => {
  const { data } = await apiClient.get(`/api/v1/user/books/homepage?userID=${userID}`);
  return data || [];
};

export const exportUserBooks = async (userID: number) => {
  try {
    // Fetch csv file from backend as a blob, treat it as a binary
    const response = await apiClient.get(`/api/v1/user/export?userID=${userID}`, {
      responseType: 'blob',
    });

    // Create blob url for file
    const blobURL = window.URL.createObjectURL(new Blob([response.data]));

    // Create invisible anchor element to trigger download
    const invisiLink = document.createElement('a');
    invisiLink.href = blobURL;
    invisiLink.setAttribute('download', 'books.csv');

    // Firefox - append link to body
    document.body.appendChild(invisiLink);

    // Trigger download by simulating click
    invisiLink.click();

    // Clean up side effect, remove link and revoke blob url
    invisiLink.parentNode?.removeChild(invisiLink);
    window.URL.revokeObjectURL(blobURL);
  } catch (error) {
    console.error("Error exporting user books: ", error);
    throw error;
  }
};



export const verifyUserToken = async () => {
  const { data } = await apiClient.get('/auth/token/verify');
  return data.user;
};

export const signOutUser = async () => {
  await apiClient.post('/auth/signout');
};

export const updateBook = async (book: Book, bookID: string) => {
  console.log('apiClient service, update book before trycatch');
  try {
    console.log('apiClient.service, updateBook, data - ', bookID);
    const { data } = await apiClient.put(`/api/v1/books/${bookID}`, book);
    return data;
  } catch (error) {
    console.error("Error updating book:", error);
    throw error;
  }
};

export const addBook = async (book: Book) => {
  console.log('apiClient.service, about to post book from addBook:', book);
  const { data } = await apiClient.post('/api/v1/books/add', book);
  console.log('apiClient.service, received response from addBook:', data);
  return data;
};

export const deleteBook = async(bookID: string) => {
  try {
    const { data } = await apiClient.delete(`/api/v1/books/${bookID}`);
    return data;
  } catch (error) {
    console.error('Error deleting book', error);
    throw error;
  }
};

export default apiClient;
