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
  notes?: string;
  formats?: ('physical' | 'eBook' | 'audioBook')[];
  createdAt?: string;
  lastUpdated?: string;
  isbn10: string;
  isbn13: string;
  isInLibrary?: boolean;
  hasEmptyFields?: boolean;
  emptyFields?: string[];
};

export type GenreData = {
  bookList: Book[];
  genreImgs: string[];
}

export type BookGenresData = {
  allGenres: string[];
  [key: string]: GenreData | string[];
}

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

export function isBookGenresData(data: any): data is BookGenresData {
  return (
    data &&
    Array.isArray(data.allGenres) &&
    Object.values(data).some(
      (value) => {
        return (
          value !== null &&
          typeof value === 'object' &&
          'bookList' in value &&
          'genreImgs' in value
        );
      }
    )
  );
}

export const defaultBookGenres: BookGenresData = {
  allGenres: [], // Correctly initialized as an array of strings
  placeholder: {
    bookList: [], // Matches `Book[]` type
    genreImgs: [], // Matches `string[]` type
  },
};