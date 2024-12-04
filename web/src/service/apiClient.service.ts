import axios, { AxiosError, InternalAxiosRequestConfig, AxiosResponse } from 'axios';
import {
  AggregatedHomePageData,
  Book,
  BookAPIPayload,
  HomePageDataResponse,
  HomepageStatistics,
  defaultBookAuthors,
  defaultHomePageStats,
  defaultBookGenres,
  defaultBookTags,
  LibraryPageResponse,
  isBookAuthorsData,
  isBookGenresData,
  isBookTagsData,
} from '../types/api';
//import { LibraryPageResponse } from '../queries/types/responses';


interface RetryConfig extends InternalAxiosRequestConfig {
  _retry?: boolean;
}

let csrfToken: string | null = null;

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_ENDPOINT,
  withCredentials: true,
  timeout: 5000
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
  response => {
    console.group(`✅ [${response.status}] Response: ${response.config.url}`);
    console.log('Response Data:', {
      status: response.status,
      statusText: response.statusText,
      headers: response.headers, // Optionally mask headers here
      data: response.data,
      timing: response.headers['x-response-time']
    });
    console.groupEnd();

    // Handle CSRF token
    const csrfTokenFromHeader = response.headers['x-csrf-token'];
    if (csrfTokenFromHeader) {
      csrfToken = csrfTokenFromHeader;
      console.log('********** interceptors.response - CRF Token capture ***********');
      console.log('CSRF token captured:', csrfToken);
      console.log('Full response headers:', response.headers);
      console.log('*********************');
    }

    // Unwrap extra data attribute from Axios response
    if (response.data &&
      typeof response.data === 'object' &&
      'data' in response.data &&
      response.config.url !== '/api/v1/csrf-token'
    ) {
      console.log('Unwrapping nested data attribute from Axios response');
      response.data = response.data.data;
    }

    return response;
  },
  // Error handler with proper typing
  async (error: AxiosError) => {
    console.group('❌ Response Error');
    console.error('Response error:', {
      message: error.message,
      status: error.response?.status,
      statusText: error.response?.statusText,
      data: error.response?.data,
      config: {
        url: error.config?.url,
        method: error.config?.method,
        params: error.config?.params
      }
    });
    console.groupEnd();

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

    if (error.code === 'ECONNABORTED' && !originalRequest._retry) {
      originalRequest._retry = true;
      console.log(`Retrying ${originalRequest.method} request to ${originalRequest.url}`);
      return apiClient(originalRequest);
    }

    return Promise.reject(error);
  }
);

apiClient.interceptors.request.use(
  async config => {
    console.group(`🚀 [${config.method?.toUpperCase()}] Request: ${config.url}`);
    console.log('Request Config:', {
      url: config.url,
      params: config.params,
      headers: config.headers, // Optionally mask sensitive headers here
      baseURL: config.baseURL
    });
    console.groupEnd();

    // For POST, PUT, DELETE requests
    if (['post', 'put', 'delete'].includes(config.method?.toLowerCase() || '')) {
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
    console.group('❌ Request Error');
    if (error.response?.status === 403 && error.response.data.includes('CSRF')) {
      console.error('CSRF Error:', error.response.data);
      console.error('Current CSRF Token:', csrfToken);
    }
    console.error('Request interceptor error:', {
      message: error.message,
      config: error.config,
      stack: error.stack
    });
    console.error('Request interceptor error:', error);
    console.groupEnd();
    return Promise.reject(error);
  }
);

// Helper function: check if homepage stats are valid
function isValidHomepageStats(stats: unknown): stats is HomepageStatistics {
  if (!stats || typeof stats !== 'object') return false;
  const s = stats as HomepageStatistics;
  return !!(
    s.userBkLang?.booksByLang &&
    s.userBkGenres?.booksByGenre &&
    s.userTags?.userTags &&
    s.userAuthors?.booksByAuthor &&
    Array.isArray(s.userBkLang.booksByLang) &&
    Array.isArray(s.userBkGenres.booksByGenre) &&
    Array.isArray(s.userTags.userTags) &&
    Array.isArray(s.userAuthors.booksByAuthor)
  );
}

// Helper function: transform API response to match library page types
function transformBookApiResponse(apiResponse: any): Book[] {
  return apiResponse.books.map((book: any) => ({
    ...book,
  }));
}

export const fetchUserBooks = async (userID: number): Promise<Book[]> => {
  console.log('apiClient.service - fetchUserBooks called');
  console.log('apiClient.service - fetchUserBooks called with userID: ', userID);

  try {
    const response = await apiClient.get('/api/v1/user/books', {
      params: { userID },
    });
    console.log('apiClient.service - fetchUserBooks full response: ', response);
    console.log('apiClient.service - fetchUserBooks status: ', response.status);
    console.log('apiClient.service - fetchUserBooks headers: ', response.headers);
    return response.data.data || [];
  } catch (error) {
    if (error instanceof AxiosError) {
      console.error('Error fetching books:', {
        message: error.message,
        status: error.response?.status,
        statusText: error.response?.statusText,
        data: error.response?.data,
        headers: error.response?.headers,
        config: {
          url: error.config?.url,
          method: error.config?.method,
          params: error.config?.params
        }
      });
    } else {
      console.error('Unknown error fetching books:', error);
    }
    throw error;
  }
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

export const fetchHomepageData = async (userID: number): Promise<HomePageDataResponse> => {
  try {
    console.log('Fetching homepage data: ', {
      userID,
      endpoint: '/api/v1/user/pages/home',
      domain: 'books'
    });

    // Request validation
    if (!userID) {
      throw new Error('UserID is required');
    }

    const response = await apiClient.get<AggregatedHomePageData>('/api/v1/pages/home', {
      params: {
        userID,
        domain: 'books'
      }
    });

    // Response validation - first check if data exists
    if (!response.data) {
      throw new Error('No data received from homepage endpoint');
    }


    // Check if books is an array
    if (!response.data?.books || !Array.isArray(response.data.books)) {
      console.error('Invalid books format:', response.data.books);
      response.data.books = [];
    }

    // Validate response
    if (!response.data || typeof response.data !== 'object') {
      console.error('Invalid response format:', response.data);
      throw new Error('Invalid response format');
    }

    if (!isValidHomepageStats(response.data.homepageStats)) {
      console.error('Invalid stats format:', response.data.homepageStats);
      response.data.homepageStats = defaultHomePageStats;
    }

    // Log response structure for debugging
    console.log('Homepage response structure: ', {
      hasBooks: !!response.data.books?.length,
      hasFormats: !!response.data.booksByFormat,
      hasStats: !!response.data.homepageStats,
      status: response.status,
    });

    // Type guard
    const defaultData: AggregatedHomePageData = {
      books: response.data.books || [],
      booksByFormat: response.data.booksByFormat || { audioBook: [], physical: [], eBook: [] },
      homepageStats: response.data.homepageStats || {
        userBkLang: { booksByLang: [] },
        userBkGenres: { booksByGenre: [] },
        userTags: { userTags: [] },
        userAuthors: { booksByAuthor: [] },
      }
    };

    // Ensure all required properties are present
    return {
      ...defaultData,
      ...response.data
    };
  } catch (error) {
    if (error instanceof AxiosError) {
      console.error('Homepage data fetch failed: ', {
        status: error.response?.status,
        statusText: error.response?.statusText,
        data: error.response?.data,
        params: error.config?.params,
        headers: error.response?.headers,
        config: {
          url: error.config?.url,
          params: error.config?.params
        }
      });
    } else {
      console.error('Unknown error fetching homepage data:', error);
    }
    throw error;
  }
};

export const fetchLibraryPageData = async (userID: number): Promise<LibraryPageResponse> => {
  try {
    console.log('Fetching library page data: ', {
      userID,
      endpoint: '/api/v1/pages/library',
      domain: 'books'
    });

    // Request validation
    if (!userID) {
      throw new Error('UserID is required');
    }

    const response: AxiosResponse<LibraryPageResponse> = await apiClient.get('/api/v1/pages/library', {
      params: {
        userID,
        domain: 'books'
      }
    });

    console.log('Raw response data:', response.data);

    // Type guard for response data
    const responseData = response.data as LibraryPageResponse;

    // Validate core data structures and provide defaults
    const validatedResponse: LibraryPageResponse = {
      booksByAuthors: isBookAuthorsData(responseData.booksByAuthors) ? responseData.booksByAuthors : defaultBookAuthors,
      booksByGenres: isBookGenresData(responseData.booksByGenres) ? responseData.booksByGenres : defaultBookGenres,
      booksByFormat: responseData.booksByFormat || { audioBook: [], eBook: [], physical: [] },
      booksByTags: isBookTagsData(responseData.booksByTags) ? responseData.booksByTags : defaultBookTags,
      books: Array.isArray(responseData.books) ? responseData.books : [],
      source: responseData.source,
      requestID: responseData.requestID
    };

    console.log('Validated response data:', validatedResponse);

    return validatedResponse;

  } catch (error) {
    if (error instanceof AxiosError) {
      console.error('Library page data fetch failed: ', {
        status: error.response?.status,
        statusText: error.response?.statusText,
        data: error.response?.data,
        params: error.config?.params,
        headers: error.response?.headers,
        config: {
          url: error.config?.url,
          params: error.config?.params
        }
      });
    } else {
      console.error('Unknown error fetching library data:', error);
    }
    throw error;
  }
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