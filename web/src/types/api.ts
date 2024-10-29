import Delta from "quill-delta";

// Individual book data structure
export type Book = {
  id?: number;
  title: string;
  subtitle?: string;
  description: string;  // Changed from optional to required
  language: string;
  pageCount: number;
  publishDate: string;  // Changed from optional to required
  authors: string[];
  imageLink: string;    // Changed from optional to required
  genres: string[];
  tags: string[];       // Changed from string[] | undefined to string[]
  notes: string | null; // Changed from optional string to string | null
  formats: ('physical' | 'eBook' | 'audioBook')[]; // Changed from optional to required
  createdAt?: string;
  lastUpdated?: string;
  isbn10?: string;
  isbn13?: string;
  isInLibrary?: boolean;
  hasEmptyFields?: boolean;
  emptyFields?: string[];
};

// Quill Delta data type used for rich text content indescription and notes fields
export type QuillContent = Delta | { ops: any[] }

// Book form data structure
export type BookFormData = {
  title: string;
  subtitle?: string;
  authors: { author: string }[]; // React Hook Form needs field array strings saved in an object
  genres: { genre: string }[]; // React Hook Form needs field array strings saved in an object
  tags: { tag: string }[]; // React Hook Form needs field array strings saved in an object
  publishDate: string;
  isbn10?: string;
  isbn13?: string;
  formats: ("physical" | "eBook" | "audioBook")[] | undefined;
  language: string;
  pageCount: number;
  imageLink: string;
  description: QuillContent; // Quill Delta object
  notes: QuillContent | null; // Quill Delta object
};

// Data for form processing
export type StringifiedBookFormData = Omit<BookFormData, 'description' | 'notes' | 'authors' | 'genres' | 'tags'> & {
  authors: string[];
  genres: string[];
  tags: string[];
  description: string;
  notes: string | null;
};

// Form data for API payload
export type BookAPIPayload = Omit<StringifiedBookFormData, 'description' | 'notes'> & {
  description: QuillContent;
  notes: QuillContent | null;
};

// Data structure for Books sorted by IndividualGenre, used on Library page for sorting
export type GenreData = {
  bookList: Book[];
  genreImgs: string[];
};

// Data structure for list of all genres, used on Library page
export type BookGenresData = {
  allGenres: string[];
  [key: string]: GenreData | string[];
};

// Data structure for Books sorted by Tag, used on Library page
export type TagData = {
  bookList: Book[];
  tagImgs: string[];
};

// Data structure for list of all tags, used on Library page
export type BookTagsData = {
  allTags: string[];
  [key: string]: TagData | string[];
};

// Book Authors/Genres intersection types
export type BookAuthorsData = {
  allAuthors: string[]
} & {
  [index: string]: Book[];
};

// Data structure for Books sorted by Format, used on Library page
export type BookFormatData = {
  audioBook: Book[];
  eBook: Book[];
  physical: Book[];
};

// Data structure for aggregated data used on Homepage for statistics
export type AggregatedHomePageData = {
  books: Book[];
  booksByFormat: BooksByFormat;
  homepageStats: HomepageStatistics;
};

/****** Type Guards */
// Type guard for Quill Delta object
export function isQuillDelta(content: any): content is Delta {
  return content && typeof content.ops !== 'undefined';
}

// Type guard for BookFormData
export function isBookFormData(book: Book | BookFormData): book is BookFormData {
  return 'authors' in book && Array.isArray(book.authors) && typeof book.authors[0] === 'object';
}

// Type guard for StringifiedBookFormData used in BookForm component
export function isStringifiedBookFormData(data: any): data is StringifiedBookFormData {
  return (
    typeof data === 'object' &&
    data !== null &&
    typeof data.title === 'string' &&
    Array.isArray(data.authors) &&
    Array.isArray(data.genres) &&
    Array.isArray(data.tags) &&
    Array.isArray(data.formats) &&
    typeof data.description === 'string'
  );
}

// Type guard for BookGenresData and BookTagsData
export function isBookData(data: any, imgKey: string) {
  return (
    data &&
    Array.isArray(data.allGenres || data.allTags) &&
    Object.values(data).some(
      (value) => {
        return (
          value !== null &&
          typeof value === 'object' &&
          'bookList' in value &&
          imgKey in value
        );
      }
    )
  );
}

// Type guard for BookGenresData
export function isBookGenresData(data: any): data is BookGenresData {
  return isBookData(data, 'genreImgs');
}

// Type guard for BookTagsData
export function isBookTagsData(data: any): data is BookTagsData {
  return isBookData(data, 'tagImgs');
}

export const defaultBookGenres: BookGenresData = {
  allGenres: [],
  placeholder: {
    bookList: [],
    genreImgs: [],
  },
};

export const defaultBookTags: BookTagsData = {
  allTags: [],
  placeholder: {
    bookList: [],
    tagImgs: [],
  },
};

/****** Utility Types */
export type TransformToStringified<T> = {
  [K in keyof T]: T[K] extends { author: string }[] ? string[] :
                  T[K] extends { genre: string }[] ? string[] :
                  T[K] extends { tag: string }[] ? string[] :
                  T[K] extends QuillContent ? string :
                  T[K];
};


export type RawHomepageStats = {
  userBkLang: Record<string, number>;
  userBkGenres: Record<string, number>;
  userTags: Record<string, number>;
  userAuthors: Record<string, number>;
};

export type BooksByFormat = {
  physical: Book[];
  eBook: Book[];
  audioBook: Book[];
};

export type HomepageStatistics = {
  userBkLang: {
    booksByLang: Array<{ label: string; count: number }>;
  };
  userBkGenres: {
    booksByGenre: Array<{ label: string; count: number }>;
  };
  userAuthors: {
    booksByAuthor: Array<{ label: string; count: number }>;
  };
  userTags: {
    userTags: Array<{ label: string; count: number }>;
  };
};

