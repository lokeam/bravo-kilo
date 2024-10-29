import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { Book, BookAPIPayload } from '../types/api';

interface RetryConfig extends InternalAxiosRequestConfig {
  _retry?: boolean;
}

let csrfToken: string | null = null;

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_ENDPOINT,
  withCredentials: true,
});

export const refreshCSRFToken = async () => {
  try {
    console.log('####');
    console.log('Attempting to refresh CSRF token');
    const response = await apiClient.get('/api/v1/csrf-token');
    console.log('CSRF token response: ', response);

    const newCSRFToken =
      response.headers['x-csrf-token'] ||
      response.headers['X-CSRF-Token'] ||
      response.headers['X-CSRF-TOKEN'];
    if (newCSRFToken) {
      csrfToken = newCSRFToken;
      apiClient.defaults.headers.common['X-CSRF-Token'] = newCSRFToken;
      console.log('New CSRF token set: ', csrfToken);
    } else {
      console.warn('No CSRF token in refresh response');
    }
  } catch (error) {
    console.error('Failed to refresh CSRF token:', error);
  }
};

apiClient.interceptors.response.use(
  // Success handler remains the same
  response => {
    const csrfTokenFromHeader = response.headers['x-csrf-token'];
    if (csrfTokenFromHeader) {
      csrfToken = csrfTokenFromHeader;
      console.log('********** interceptors.response - CRF Token capture ***********');
      console.log('CSRF token captured:', csrfToken);
      console.log('Full response headers:', response.headers);
      console.log('*********************');
    }
    return response;
  },
  // Error handler with proper typing
  async (error: AxiosError) => {
    console.error('Response error:', error);
    const originalRequest = error.config as RetryConfig;

    if (!originalRequest) {
      return Promise.reject(error);
    }

    // Extract the URL path
    const urlPath = new URL(originalRequest.url || '', apiClient.defaults.baseURL).pathname;

    // Exclude specific URLs from interceptor logic
    const excludedUrls = ['/auth/token/verify', '/auth/token/refresh', '/auth/google/signin'];
    if (excludedUrls.includes(urlPath)) {
      return Promise.reject(error);
    }

    // Check if error response and status exist
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        console.log('apiClient, trying token refresh');
        await apiClient.post('/auth/token/refresh');
        return apiClient(originalRequest);
      } catch (refreshError) {
        console.error('Token refresh failed', refreshError);
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }

    // If it's a CSRF error, try to refresh the token and retry the request
    if (error.response?.status === 403) {
      try {
        await refreshCSRFToken();
        return apiClient(originalRequest);
      } catch (csrfError) {
        console.error('CSRF refresh failed:', csrfError);
        return Promise.reject(csrfError);
      }
    }

    return Promise.reject(error);
  }
);

apiClient.interceptors.request.use(
  async config => {
    // For POST, PUT, DELETE requests
    if (['post', 'put', 'delete'].includes(config.method?.toLowerCase() || '')) {
      // If we don't have a CSRF token, try to get one
      if (!csrfToken) {
        await refreshCSRFToken();
      }

      // Set the CSRF token in the request headers
      if (csrfToken) {
        config.headers['X-CSRF-Token'] = csrfToken;
        console.log('********** interceptors.request - CSRF Token capture ***********');
        console.log(`CSRF token added to ${config.method?.toUpperCase()} request:`, csrfToken);
      } else {
        console.warn(`No CSRF token available for ${config.method?.toUpperCase()} request`);
      }
      console.log('Full request config:', config);
      console.log('*********************');
    }
    return config;
  },
  error => {
    if (error.response && error.response.status === 403 && error.response.data.includes('CSRF')) {
      console.error('CSRF Error:', error.response.data);
      console.error('Current CSRF Token:', csrfToken);
    }
    console.error('Request interceptor error:', error);
    return Promise.reject(error);
  }
);

export const fetchUserBooks = async (userID?: number) => {
  if (userID === undefined) return [];
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

export const fetchBooksTags = async (userID: number) => {
  console.log('=====');
  console.log('fetchBooksTags called within apiClient service');
  const { data } = await apiClient.get(`/api/v1/user/books/tags?user=${userID}`);
  console.log('fetchBooksTags, data: ', data);
  return data || {};
}

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
  try {
    const { data } = await apiClient.get(`/api/v1/user/books/homepage?userID=${userID}`);
    return data || {
      userBkLang: { booksByLang: [] },
      userBkGenres: { booksByGenre: [] },
      userAuthors: { booksByAuthor: [] },
      userTags: { userTags: [] }
    };
  } catch (error) {
    console.error('Error fetching homepage data:', error);
  }

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

export const updateBook = async (book: BookAPIPayload, bookID: string): Promise<Book> => {
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

export const addBook = async (book: BookAPIPayload): Promise<Book> => {
  console.log('apiClient.service, received book data:', book);

  console.log('Description:', book.description);
  console.log('Notes:', book.notes);

  if (!book.description) {
    console.warn('Warning: Description is empty or null');
  }
  if (!book.notes) {
    console.warn('Warning: Notes is empty or null');
  }

  try {
    const { data } = await apiClient.post('/api/v1/books/add', book);
    console.log('apiClient.service, received response from addBook:', data);
    return data;
  } catch (error) {
    console.error('Error adding book:', error);
    throw error;
  }
};
//
export const checkUserAccountStatus = async () => {
  try {
    const { data } = await apiClient.get('/auth/check-account-status');
    return data;
  } catch (error) {
    console.error('Error checking account status', error);
    throw error;
  }
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

export const deleteUser = async() => {
  try {
    console.log('--------------------------------');
    console.log('apiClient.service, about to delete user');
    const { data } = await apiClient.delete(`/auth/delete-account`);
    console.log('apiClient.service, user deleted:', data);
    console.log('--------------------------------');
    return data;
  } catch {
    console.log('Error attempting to delete user');
  }
};

export const initiateAccountRecovery = async () => {
  try {
    console.log('apiClient.service, about to initiate account recovery');
    const { data } = await apiClient.post('/auth/account-recovery');
    console.log('apiClient.service, account recovery initiated: ', data);
    return { success: true, message: data.message };
  } catch (error) {
    console.error('Error during account recovery');
    if (
      error instanceof AxiosError &&
      error.response?.data?.message
    ) {
      return { success: false, message: error.response?.data?.message };
    }
    return { success: false, message: 'An unknown error occurred' };
  }
};

export default apiClient;