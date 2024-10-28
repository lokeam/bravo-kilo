import { Book, BookFormData, StringifiedBookFormData } from '../types/api';
import Delta from 'quill-delta';

export const transformBookData = (bookData: Partial<Book> = {}, formattedDate: string): BookFormData => {
  const parseField = (field: string | undefined | null): Delta => {
    if (!field) return new Delta();
    try {
      const parsed = JSON.parse(field);
      return new Delta(Array.isArray(parsed.ops) ? parsed : { ops: [{ insert: field }] });
    } catch (error) {
      console.log('Creating Delta object from plain text:', field);
      return new Delta().insert(field);
    }
  };

  return {
    title: bookData.title || '',
    subtitle: bookData.subtitle || '',
    authors: bookData.authors
      ? bookData.authors.map((author: string) => ({ author }))
      : [{ author: '' }],
    genres: bookData.genres
      ? bookData.genres.map((genre: string) => ({ genre }))
      : [{ genre: '' }],
    tags: bookData.tags
      ? bookData.tags.map((tag: string) => ({ tag }))
      : [{ tag: '' }],
    publishDate: formattedDate,
    isbn10: bookData.isbn10 || '',
    isbn13: bookData.isbn13 || '',
    formats: bookData.formats || [],
    language: bookData.language || 'en',
    pageCount: bookData.pageCount || 0,
    imageLink: bookData.imageLink || '',
    description: parseField(bookData.description),
    notes: parseField(bookData.notes),
  };
};

export const transformFormData = (data: BookFormData): StringifiedBookFormData => {
  return {
    ...data,
    authors: data.authors.map(a => a.author.trim()).filter(author => author !== ''),
    genres: data.genres.map(g => g.genre.trim()).filter(genre => genre !== ''),
    tags: data.tags.map(t => t.tag.trim()).filter(tag => tag !== ''),
    description: JSON.stringify(data.description),
    notes: data.notes ? JSON.stringify(data.notes) : null,
  };
};