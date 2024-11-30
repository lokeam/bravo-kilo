export const scrollbarWidth = window.innerWidth - document.documentElement.clientWidth;
import { Book } from '../types/api';

/****** Library Page Sorting Function ******/
export const sortBooks = (
  books: Book[],
  sortCriteria: string,
  sortOrder: 'asc' | 'desc'
): Book[] => {
  return [...books].sort((a, b) => {
    switch (sortCriteria) {
      case 'title':
        return sortOrder === 'asc' ? a.title.localeCompare(b.title) : b.title.localeCompare(a.title);
      case 'publishDate': {
        const dateA = a.publishDate ? new Date(a.publishDate).getTime() : 0;
        const dateB = b.publishDate ? new Date(b.publishDate).getTime() : 0;
        return sortOrder === 'asc' ? dateA - dateB : dateB - dateA;
      }
      case 'author': {
        const aSurname = a.authors?.[0]?.split(' ').pop() || '';
        const bSurname = b.authors?.[0]?.split(' ').pop() || '';
        return sortOrder === 'asc' ? aSurname.localeCompare(bSurname) : bSurname.localeCompare(aSurname);
      }
      case 'pageCount':
        return sortOrder === 'asc' ? a.pageCount - b.pageCount : b.pageCount - a.pageCount;
      default:
        return 0;
    }
  });
};