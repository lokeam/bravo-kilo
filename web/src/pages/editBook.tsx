import { useEffect } from 'react';
import { useForm, SubmitHandler } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Book } from './library';
import axios from 'axios';

const fetchBook = async (bookID: string) => {
  console.log('-------');
  console.log(`Fetching book with ID: ${bookID}`);
  const { data } = await axios.get(`${import.meta.env.VITE_API_ENDPOINT}/api/v1/books/${bookID}`, {
    withCredentials: true,
  });
  return data.book;
};

const bookSchema = z.object({
  id: z.number(),
  title: z.string().min(1, 'Please enter a title'),
  subtitle: z.string().optional(),
  description: z.string().min(1, 'Please enter a description'),
  language: z.string().min(1, 'Please enter a language'),
  pageCount: z.number().min(1, 'Please enter a total page count'),
  publishDate: z.string().min(1, 'Please enter a date of publication'),
  authors: z.array(z.string()).min(1, 'Please enter at least one author'),
  imageLinks: z.array(z.string()).min(1, 'At least one image link is required'),
  genres: z.array(z.string()).min(1, 'Please enter at least one genre'),
  tags: z.array(z.string()).min(1, 'At least one tag is required'),
  notes: z.string().optional(),
  formats: z.array(z.enum(['physical', 'eBook', 'audioBook'])),
  isbn10: z.string().min(10).max(10),
  isbn13: z.string().min(13).max(13),
});

type BookFormData = z.infer<typeof bookSchema>;

const EditBook = () => {
  const { bookID } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: book, isLoading, isError } = useQuery({
    queryKey: ['book', bookID],
    queryFn: () => fetchBook(bookID as string),
    enabled: !!bookID,
  });

  const updateBook = async (book: Book) => {
    console.log(`Updating book with data: ${book}`);
    const { data } = await axios.put(`${import.meta.env.VITE_API_ENDPOINT}/api/v1/books/${bookID}`, book, {
      withCredentials: true,
    });
    console.log(`Update response data: ${data}`);
    return data;
  };

  const { register, handleSubmit, reset, formState: { errors } } = useForm<BookFormData>({
    resolver: zodResolver(bookSchema),
  });

  const mutation = useMutation({
    mutationFn: updateBook,
    onSuccess: () => {
      console.log('Book updated successfully');
      queryClient.invalidateQueries({ queryKey: ['book', bookID] });
      navigate(`/library/books/${bookID}`);
    },
    onError: (error) => {
      console.error(`Error updating book: ${error}`);
    }
  });

  useEffect(() => {
    if (book) reset(book);

    console.log(`useEffect fetched book data: ${book}`);
  }, [book, reset]);

  if (isLoading) return <div>Loading...</div>;

  if (isError) return <div>Error loading book data</div>;

  const onSubmit: SubmitHandler<BookFormData> = (data) => {
    console.log(`Form submitted with data ${data}`);
    const defaultDate = new Date().toISOString();

    const book: Book = {
      ...data,
      id: Number(data.id),
      createdAt: defaultDate,
      lastUpdated: defaultDate,
    };

    mutation.mutate(book);
  };

  return (
    <section>
      <div>
        <h2>Edit Book</h2>
        <form onSubmit={handleSubmit(onSubmit)}>
          <div>
            <label htmlFor="title">Title</label>
            <input id="title" {...register('title')} />
          </div>
          <div>
            <label htmlFor="subtitle">Subtitle</label>
            <input id="subtitle" {...register('subtitle')} />
          </div>
          <div>
            <label htmlFor="description">Description</label>
            <textarea id="description" {...register('description')} />
          </div>
          <div>
            <label htmlFor="language">Language</label>
            <input id="language" {...register('language')} />
          </div>
          <div>
            <label htmlFor="pageCount">Page Count</label>
            <input id="pageCount" type="number" {...register('pageCount')} />
          </div>
          <div>
            <label htmlFor="publishDate">Publish Date</label>
            <input id="publishDate" {...register('publishDate')} />
          </div>
          <div>
            <label htmlFor="authors">Authors</label>
            <input id="authors" {...register('authors')} />
          </div>
          <div>
            <label htmlFor="imageLinks">Image Links</label>
            <input id="imageLinks" {...register('imageLinks')} />
          </div>
          <div>
            <label htmlFor="genres">Genres</label>
            <input id="genres" {...register('genres')} />
          </div>
          <div>
            <label htmlFor="tags">Tags</label>
            <input id="tags" {...register('tags')} />
          </div>
          <div>
            <label htmlFor="notes">Notes</label>
            <textarea id="notes" {...register('notes')} />
          </div>
          <div>
            <label htmlFor="formats">Formats</label>
            <select id="formats" {...register('formats')} multiple>
              <option value="physical">Physical</option>
              <option value="eBook">eBook</option>
              <option value="audioBook">AudioBook</option>
            </select>
          </div>
          <div>
            <label htmlFor="isbn10">ISBN-10</label>
            <input id="isbn10" {...register('isbn10')} />
          </div>
          <div>
            <label htmlFor="isbn13">ISBN-13</label>
            <input id="isbn13" {...register('isbn13')} />
          </div>
          <button type="submit">Update Book</button>
        </form>
      </div>
    </section>
  );
};

export default EditBook;
