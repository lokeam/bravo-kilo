import { z } from 'zod';
import Delta from 'quill-delta';

const deltaSchema = z.custom<Delta>(
  (data) => data instanceof Delta && isValidDelta(data),
  { message: 'Content cannot be empty' }
);

const isValidDelta = (data: any) => {
  return data.ops && data.ops.length > 0 && data.ops.some((op: any) =>
    (typeof op.insert === 'string' && op.insert.trim() !== '') ||
    (typeof op.insert === 'object' && Object.keys(op.insert).length > 0)
  );
};

export const bookSchema = z.object({
  title: z.string().min(1, 'Please enter a title'),
  subtitle: z.string().optional(),
  authors: z
    .array(
      z.object({
        author: z.string().min(1, 'Author name cannot be empty'),
      })
    )
    .min(1, 'Please enter at least one author'),
  genres: z
    .array(
      z.object({
        genre: z.string().min(1, 'Genre cannot be empty'),
      })
    )
    .min(1, 'Please enter at least one genre'),
  tags: z
    .array(
      z.object({
        tag: z.string().min(1, 'Tag cannot be empty'),
      })
    )
    .min(1, 'At least one tag is required'),
  publishDate: z.string().min(1, 'Please enter a date of publication'),
  isbn10: z.string().length(10, 'ISBN-10 must be 10 characters').optional(),
  isbn13: z.string().length(13, 'ISBN-13 must be 13 characters').optional(),
  formats: z
    .array(z.enum(['physical', 'eBook', 'audioBook']))
    .min(1, 'Please select at least one format'),
  language: z.string().min(1, 'Please enter a language'),
  pageCount: z.number().min(1, 'Please enter a total page count'),
  imageLink: z.string().min(1, 'Please enter an image link'),
  description: deltaSchema,
  notes: deltaSchema.nullable(),
}).refine((data) => data.isbn10 || data.isbn13, {
  message: 'Either ISBN-10 or ISBN-13 is required',
  path: ['isbn10', 'isbn13'],
});
