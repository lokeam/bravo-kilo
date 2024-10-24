import Delta from "quill-delta";

export type Book = {
  id?: number;
  title: string;
  subtitle?: string;
  description?: string;
  language: string;
  pageCount: number;
  publishDate?: string;
  authors: string[];
  imageLink?: string;
  genres: string[];
  tags?: string[] | undefined;
  notes?: string;
  formats?: ('physical' | 'eBook' | 'audioBook')[];
  createdAt?: string;
  lastUpdated?: string;
  isbn10?: string;
  isbn13?: string;
  isInLibrary?: boolean;
  hasEmptyFields?: boolean;
  emptyFields?: string[];
};

export type GenreData = {
  bookList: Book[];
  genreImgs: string[];
};

export type BookGenresData = {
  allGenres: string[];
  [key: string]: GenreData | string[];
};

export type TagData = {
  bookList: Book[];
  tagImgs: string[];
};

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

export type BookFormatData = {
  audioBook: Book[];
  eBook: Book[];
  physical: Book[];
};

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

export function isBookGenresData(data: any): data is BookGenresData {
  return isBookData(data, 'genreImgs');
}

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

export type AggregatedHomePageData = {
  books: Book[];
  booksByFormat: BooksByFormat;
  homepageStats: HomepageStatistics;
};

export type BookFormData = {
  title: string;
  subtitle?: string;
  authors: { author: string }[]; // React Hook Form needs field array strings saved in an object
  genres: { genre: string }[];
  tags: { tag: string }[];
  publishDate: string;
  isbn10: string;
  isbn13: string;
  formats: ("physical" | "eBook" | "audioBook")[] | undefined;
  language: string;
  pageCount: number;
  imageLink: string;
  description: Delta;
  notes: Delta | null;
};
