import { useEffect } from 'react';
import { Controller, SubmitHandler, useForm, useFieldArray } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate, useLocation } from 'react-router-dom';

import useAddBook from '../hooks/useAddBook';
import { Book } from './Library';

import { IoClose } from 'react-icons/io5';
import { IoAddOutline } from 'react-icons/io5';


const bookSchema = z.object({
  title: z.string().min(1, 'Please enter a title'),
  subtitle: z.string().optional(),
  authors: z.array(z.string()).min(1, 'Please enter at least one author'),
  genres: z.array(z.string()).min(1, 'Please enter at least one genre'),
  tags: z.array(z.string()).min(1, 'At least one tag is required'),
  publishDate: z.string().min(1, 'Please enter a date of publication'),
  isbn10: z.string().min(10).max(10),
  isbn13: z.string().min(13).max(13),
  formats: z.array(z.enum(['physical', 'eBook', 'audioBook'])),
  language: z.string().min(1, 'Please enter a language'),
  pageCount: z.number().min(1, 'Please enter a total page count'),
  imageLinks: z.array(z.string()).min(1, 'At least one image link is required'),
  description: z.string().min(1, 'Please enter a description'),
  notes: z.string().optional(),
});

type BookFormData = z.infer<typeof bookSchema>;

const ManualAdd = () => {
  const { mutate: addBook } = useAddBook();
  const navigate = useNavigate();
  const location = useLocation();

  // Retrieve book data from navigation state
  const bookData = location.state?.book || {};

  const {
    control,
    handleSubmit,
    register,
    reset,
    formState: { errors }
  } = useForm<BookFormData>({
    resolver: zodResolver(bookSchema),
    defaultValues: bookData, // Set default values from the state
  });

  const {
    fields: authorFields,
    append: appendAuthor,
    remove: removeAuthor
  } = useFieldArray({
    control,
    name: 'authors' as const,
  });

  const {
    fields: genreFields,
    append: appendGenre,
    remove: removeGenre
  } = useFieldArray({
    control,
    name: 'genres' as const,
  });

  const {
    fields: tagFields,
    append: appendTag,
    remove: removeTag
  } = useFieldArray({
    control,
    name: 'tags' as const,
  });

  const {
    fields: imageLinkFields,
    append: appendImageLink,
    remove: removeImageLink
  } = useFieldArray({
    control,
    name: 'imageLinks' as const,
  });

  useEffect(() => {
    if (bookData) reset(bookData); // Reset form fields with book data
  }, [bookData, reset]);

  const onSubmit: SubmitHandler<BookFormData> = (data) => {
    const defaultDate = new Date().toISOString();

    const book: Book = {
      ...data,
      subtitle: data.subtitle || '',
      imageLinks: data.imageLinks.map(link => link.trim()), // Ensure imageLinks is an array of trimmed strings
      createdAt: defaultDate,
      lastUpdated: defaultDate,
    };

    addBook(book);
    navigate('/library/');
  };

  return (
<section className="bg-white dark:bg-gray-900">
      <div className="text-left py-8 px-4 mx-auto max-w-2xl lg:py-16">
        <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">Add Book</h2>
        <form onSubmit={handleSubmit(onSubmit)}>

          {/* Title */}
          <div>
            <label htmlFor="title">Title</label>
            <input
              id="title"
              type="text"
              {...register("title")}
              className={`border ${errors.title ? 'border-red-500' : 'border-gray-300'}`}
            />
            {errors.title && <p className="text-red-500">{errors.title.message}</p>}
          </div>

          {/* Subtitle */}
          <div className="block sm:col-span-2">
            <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white" htmlFor="subtitle">Subtitle</label>
            <input
              id="subtitle"
              className="bg-gray-50 border border-gray-00 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500"
              {...register('subtitle')}
            />
          </div>

          {/* Authors Field Array */}
          {authorFields.map((item, index) => (
            <div className="flex items-center p-4" key={item.id}>
              <Controller
                render={({ field }) => <input {...field} className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />}
                name={`authors.${index}`}
                control={control}
              />
              <button type="button" onClick={() => authorFields.length > 1 && removeAuthor(index)}  className="ml-5 rounded bg-transparent">
                <IoClose size={20}/>
              </button>
            </div>
          ))}
          <button type="button" onClick={() => appendAuthor('')} className="flex flex-row justify-between items-center bg-transparent m-4">
            <IoAddOutline size={20} className="mr-1"/>
            Add Author
          </button>

          {/* Genres Field Array */}
          {genreFields.map((item, index) => (
            <div key={item.id} className="flex items-center p-4">
              <Controller
                render={({ field }) => <input {...field} className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500"/>}
                name={`genres.${index}`}
                control={control}
              />
              <button type="button" onClick={() => genreFields.length > 1 && removeGenre(index)} className="ml-5">
                <IoClose size={20}/>
              </button>
            </div>
          ))}
          <button type="button" onClick={() => appendGenre('')} className="flex flex-row justify-between items-center bg-transparent m-4">
            <IoAddOutline size={20} className="mr-1"/>
            Add Genre
          </button>

          {/* Tags Field Array */}
          {tagFields.map((item, index) => (
            <div key={item.id} className="flex items-center p-4">
              <Controller
                render={({ field }) => <input {...field} className="bg-gray-50 text-gray-900 text-sm focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />}
                name={`tags.${index}`}
                control={control}
              />
              <button type="button" onClick={() => tagFields.length > 1 && removeTag(index)} className="flex flex-row justify-between items-center bg-transparent mt-2 ml-5">
                <IoClose size={20}/>
              </button>
            </div>
          ))}
          <button type="button" onClick={() => appendTag('')} className="flex flex-row justify-between items-center bg-transparent mt-2">
            <IoAddOutline size={20} className="mr-1"/>
            Add Tag
          </button>
{/* Publish Date */}
          <div>
            <label htmlFor="publishDate">Publish Date</label>
            <input
              id="publishDate"
              type="text"
              {...register("publishDate")}
              className={`border ${errors.publishDate ? 'border-red-500' : 'border-gray-300'}`}
            />
            {errors.publishDate && <p className="text-red-500">{errors.publishDate.message}</p>}
          </div>

          {/* ISBN10 */}
          <div>
            <label htmlFor="isbn10">ISBN-10</label>
            <input
              id="isbn10"
              type="text"
              {...register("isbn10")}
              className={`border ${errors.isbn10 ? 'border-red-500' : 'border-gray-300'}`}
            />
            {errors.isbn10 && <p className="text-red-500">{errors.isbn10.message}</p>}
          </div>

          {/* ISBN13 */}
          <div>
            <label htmlFor="isbn13">ISBN-13</label>
            <input
              id="isbn13"
              type="text"
              {...register("isbn13")}
              className={`border ${errors.isbn13 ? 'border-red-500' : 'border-gray-300'}`}
            />
            {errors.isbn13 && <p className="text-red-500">{errors.isbn13.message}</p>}
          </div>

          {/* Formats */}
          <div>
            <label>Formats</label>
            <ul className="grid w-full gap-6 md:grid-cols-3">
              {['physical', 'eBook', 'audioBook'].map((format) => (
                <li key={format}>
                  <input
                    type="checkbox"
                    id={`formats_${format}`}
                    {...register('formats')}
                    value={format}
                    className="hidden peer"
                  />
                  <label htmlFor={`formats_${format}`} className="inline-flex text-center items-center justify-center w-full p-2 text-gray-500 bg-white border-2 border-gray-200 rounded cursor-pointer dark:hover:text-gray-300 dark:border-gray-700 peer-checked:border-blue-600 hover:text-gray-600 dark:peer-checked:text-gray-300 peer-checked:text-gray-600 hover:bg-gray-50 dark:text-gray-400 dark:bg-gray-800 dark:hover:bg-gray-700">{format}</label>
                </li>
              ))}
            </ul>
          </div>

          {/* Language */}
          <div>
            <label htmlFor="language">Language</label>
            <input
              id="language"
              type="text"
              {...register("language")}
              className={`border ${errors.language ? 'border-red-500' : 'border-gray-300'}`}
            />
            {errors.language && <p className="text-red-500">{errors.language.message}</p>}
          </div>

          {/* Page Count */}
          <div>
            <label htmlFor="pageCount">Page Count</label>
            <input
              id="pageCount"
              type="number"
              {...register("pageCount", { valueAsNumber: true })}
              className={`border ${errors.pageCount ? 'border-red-500' : 'border-gray-300'}`}
            />
            {errors.pageCount && <p className="text-red-500">{errors.pageCount.message}</p>}
          </div>

          {/* Image Links Field Array */}
          {imageLinkFields.map((item, index) => (
            <div key={item.id} className="flex items-center p-4">
              <Controller
                render={({ field }) => <input {...field} className="bg-gray-50 text-gray-900 text-sm focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />}
                name={`imageLinks.${index}`}
                control={control}
              />
              <button type="button" onClick={() => imageLinkFields.length > 1 && removeImageLink(index)} className="flex flex-row justify-between items-center bg-transparent mt-2 ml-5">
                <IoClose size={20}/>
              </button>
            </div>
          ))}
          <button type="button" onClick={() => appendImageLink('')} className="flex flex-row justify-between items-center bg-transparent mt-2">
            <IoAddOutline size={20} className="mr-1"/>
            Add Image Link
          </button>

          {/* Description */}
          <div>
            <label htmlFor="description">Description</label>
            <textarea
              id="description"
              rows={4}
              {...register("description")}
              className={`border ${errors.description ? 'border-red-500' : 'border-gray-300'}`}
            />
            {errors.description && <p className="text-red-500">{errors.description.message}</p>}
          </div>

          {/* Notes */}
          <div>
            <label htmlFor="notes">Notes</label>
            <textarea
              id="notes"
              rows={4}
              {...register("notes")}
              className="border border-gray-300"
            />
          </div>

          <button type="submit" className="mt-4 p-2 bg-blue-500 text-white rounded">
            Add Book
          </button>
        </form>
      </div>
    </section>
  );
};

export default ManualAdd;