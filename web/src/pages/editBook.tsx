import { useEffect } from 'react';
import { Controller, useForm, useFieldArray, SubmitHandler } from 'react-hook-form';
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

  const { register, handleSubmit, control, reset, formState: { errors } } = useForm<BookFormData>({
    resolver: zodResolver(bookSchema),
  });

  const { fields: authorFields, append: appendAuthor, remove: removeAuthor } = useFieldArray({
    control,
    name: 'authors' as const,
  });

  const { fields: genreFields, append: appendGenre, remove: removeGenre } = useFieldArray({
    control,
    name: 'genres' as const,
  });

  const { fields: tagFields, append: appendTag, remove: removeTag } = useFieldArray({
    control,
    name: 'tags' as const,
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

          {/* Authors Field Array */}
          <div>
            <label>Authors</label>
            {authorFields.map((item, index) => (
              <div key={item.id}>
                <Controller
                  render={({ field }) => <input {...field} />}
                  name={`authors.${index}`}
                  control={control}
                />
                <button type="button" onClick={() => authorFields.length > 1 && removeAuthor(index)}>
                  Delete
                </button>
              </div>
            ))}
            <button type="button" onClick={() => appendAuthor('')}>
              Add Author
            </button>
          </div>

          {/* Genres Field Array */}
          <div>
            <label>Genres</label>
              {genreFields.map((item, index) => (
                <div key={item.id}>
                  <Controller
                    render={({ field }) => <input {...field} />}
                    name={`genres.${index}`}
                    control={control}
                  />
                  <button type="button" onClick={() => genreFields.length > 1 && removeGenre(index)}>
                    Delete
                  </button>
                </div>
              ))}
              <button type="button" onClick={() => appendGenre('')}>
                Add Genre
              </button>
          </div>

          {/* Tags Field Array */}
          <div>
          {tagFields.map((item, index) => (
              <div key={item.id}>
                <Controller
                  render={({ field }) => <input {...field} />}
                  name={`tags.${index}`}
                  control={control}
                />
                <button type="button" onClick={() => tagFields.length > 1 && removeTag(index)}>
                  Delete
                </button>
              </div>
            ))}
            <button type="button" onClick={() => appendTag('')}>
              Add Tag
            </button>
          </div>

          <div>
            <label htmlFor="publishDate">Publish Date</label>
            <input id="publishDate" {...register('publishDate')} />
          </div>

          <div>
            <label htmlFor="isbn10">ISBN-10</label>
            <input id="isbn10" {...register('isbn10')} />
          </div>

          <div>
            <label htmlFor="isbn13">ISBN-13</label>
            <input id="isbn13" {...register('isbn13')} />
          </div>

          {/* Formats */}
          <div>
            <label>Formats</label>
            {['physical', 'eBook', 'audioBook'].map((format) => (
              <div key={format}>
                <label htmlFor={`formats_${format}`}>
                  <input
                    type="checkbox"
                    id={`formats_${format}`}
                    {...register('formats')}
                    value={format}
                  />
                  {format}
                </label>
              </div>
            ))}
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
            <label htmlFor="imageLinks">Image Links</label>
            <input id="imageLinks" {...register('imageLinks')} />
          </div>

          <div>
            <label htmlFor="description">Description</label>
            <textarea id="description" {...register('description')} />
          </div>

          <div>
            <label htmlFor="notes">Notes</label>
            <textarea id="notes" {...register('notes')} />
          </div>

          <button type="submit">Update Book</button>
        </form>
      </div>
    </section>
  );
};

export default EditBook;
