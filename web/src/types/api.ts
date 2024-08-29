export type Book = {
  id?: number;
  title: string;
  subtitle?: string;
  description?: string;
  language: string;
  pageCount: number;
  publishDate?: string;
  authors: string[];
  imageLink: string;
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