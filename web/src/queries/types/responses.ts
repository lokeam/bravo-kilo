import type { Book, BookGenresData, BookTagsData, BookAuthorsData, BookFormatData } from '../../types/api';

// export interface LibraryPageData {
//   books: Book[];
//   booksByAuthors: BookAuthorsData;
//   booksByGenres: BookGenresData;
//   booksByFormat: BookFormatData;
//   booksByTags: BookTagsData;
// }

// export interface LibraryPageResponse {
//   requestId: string;
//   data: LibraryPageData;
//   source: 'cache' | 'database';
// }

export interface APIErrorResponse {
  error: string;
}