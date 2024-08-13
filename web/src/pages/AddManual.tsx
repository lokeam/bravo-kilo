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

// Use default values to ensure inputs are always controlled
const {
  control,
  handleSubmit,
  register,
  reset,
  formState: { errors },
} = useForm<BookFormData>({
  resolver: zodResolver(bookSchema),
  defaultValues: {
    title: '',
    subtitle: '',
    authors: [''],
    genres: ['Genre 1'],
    tags: [''],
    publishDate: '',
    isbn10: '',
    isbn13: '',
    formats: [],
    language: '',
    pageCount: 0,
    imageLinks: [''],
    description: '',
    notes: '',
    ...bookData, // Merge with any existing data
  },
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
    // Only reset if bookData is actually different
    if (bookData && Object.keys(bookData).length > 0) {
      reset(bookData);
    }
  }, [bookData, reset]);

  const onSubmit: SubmitHandler<BookFormData> = (data) => {
    console.log('Submitted data:', data);
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

  console.log('RHF errors: ', errors );

  return (
    <section className="bg-white dark:bg-gray-900 my-20">
      <div className="text-left px-4 py-8 md:pl-24 mx-auto max-w-2xl lg:py-16">
        <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">Add Book</h2>
        <form className="grid gap-4 sm:grid-cols-2 sm:gap-6" onSubmit={handleSubmit(onSubmit)}>

          {/* Title */}
          <div className="sm:col-span-2">
            <label
              className="block mb-2 text-sm font-medium text-gray-900 dark:text-white"
              id="title"
              htmlFor="title"
            >
              Title
              <span className="text-red-600 ml-px">*</span>
            </label>
            <input
              id="title"
              type="text"
              placeholder="Enter a book title"
              {...register("title")}
              className={`
                border ${errors.title ? 'border-red-500' : 'border-gray-300'}
                bg-gray-50 border border-gray-00 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500
                `}
            />
            {errors.title && <p className="text-red-500">{errors.title.message}</p>}
          </div>

          {/* Subtitle */}
          <div className="block sm:col-span-2">
            <label
              className="block mb-2 text-sm font-medium text-gray-900 dark:text-white"
              htmlFor="subtitle"
              >
                Subtitle
              </label>
            <input
              id="subtitle"
              className="bg-gray-50 border border-gray-00 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500"
              placeholder="Subtitle optional"
              {...register('subtitle')}
            />
            {errors.subtitle && <p className="text-red-500">{errors.subtitle.message}</p>}
          </div>

          {/* Authors Field Array */}
          <div className="block sm:col-span-2">
            <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Authors<span className="text-red-600 ml-px">*</span></label>
            <div className="border border-gray-300 rounded">
              {authorFields.map((item, index) => (
                <div className="flex items-center p-4" key={item.id}>
                  <Controller
                    render={({ field }) => <input {...field} className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />}
                    name={`authors.${index}`}
                    control={control}
                  />
                  <button type="button" onClick={() => authorFields.length > 1 && removeAuthor(index)}  className="ml-5 rounded bg-dark-clay">
                    <IoClose size={20}/>
                  </button>
                </div>
              ))}
              <button type="button" onClick={() => appendAuthor('')} className="flex flex-row justify-between items-center bg-dark-clay m-4">
                <IoAddOutline size={20} className="mr-1"/>
                Add Author
              </button>
            </div>
            {errors.authors && <p className="text-red-500">{errors.authors.message}</p>}
          </div>


          {/* Genres Field Array */}
          <div className="block sm:col-span-2">
            <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Genres<span className="text-red-600 ml-px">*</span></label>
            <div className="border border-gray-300 rounded">
                {genreFields.map((item, index) => (
                  <div key={item.id} className="flex items-center p-4">
                    <Controller
                      render={({ field }) => <input {...field} className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500"/>}
                      name={`genres.${index}`}
                      control={control}
                    />
                    <button type="button" onClick={() => genreFields.length > 1 && removeGenre(index)} className="ml-5 bg-dark-clay">
                      <IoClose size={20}/>
                    </button>
                  </div>
                ))}
              <button type="button" onClick={() => appendGenre('')} className="flex flex-row justify-between items-center bg-dark-clay m-4">
                <IoAddOutline size={20} className="mr-1"/>
                Add Genre
              </button>
            </div>
            {errors.genres && <p className="text-red-500">{errors.genres.message}</p>}
          </div>

          {/* Tags Field Array */}
          <div className="block sm:col-span-2">
            <label htmlFor="tags" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Personal Tags<span className="text-red-600 ml-px">*</span></label>
            <div className="border border-gray-300 rounded">
              {tagFields.map((item, index) => (
                  <div key={item.id} className="flex items-center p-4">
                    <Controller
                      render={({ field }) => <input {...field} className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />}
                      name={`tags.${index}`}
                      control={control}
                    />
                    <button type="button" onClick={() => tagFields.length > 1 && removeTag(index)} className="flex flex-row justify-between items-center bg-dark-clay m-4">
                      <IoClose size={20}/>
                    </button>
                  </div>
                ))}
                <button type="button" onClick={() => appendTag('')} className="flex flex-row justify-between items-center bg-dark-clay m-4">
                  <IoAddOutline size={20} className="mr-1"/>
                  Add Tag
                </button>
            </div>
            {errors.tags && <p className="text-red-500">Please enter at least one personal tag</p>}
          </div>


          {/* Publish Date */}
          <div className="sm:col-span-2">
            <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white" htmlFor="publishDate">Publish Date<span className="text-red-600 ml-px">*</span></label>
            <input
              id="publishDate"
              placeholder="YYYY-MM-DD"
              type="text"
              {...register("publishDate")}
              className={`
                border ${errors.publishDate ? 'border-red-500' : 'border-gray-300'}
                bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500
                `}
            />
            {errors.publishDate && <p className="text-red-500">{errors.publishDate.message}</p>}
          </div>

          {/* ISBN10 */}
          <div className="w-full">
            <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white" htmlFor="isbn10">ISBN-10<span className="text-red-600 ml-px">*</span></label>
            <input
              id="isbn10"
              placeholder="XXXXXXXXXX"
              type="text"
              {...register("isbn10")}
              className={`
                border ${errors.isbn10 ? 'border-red-500' : 'border-gray-300'}
                bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500
                `}
            />
            {errors.isbn10 && <p className="text-red-500">{errors.isbn10.message}</p>}
          </div>

          {/* ISBN13 */}
          <div className="w-full">
            <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white" htmlFor="isbn13">ISBN-13<span className="text-red-600 ml-px">*</span></label>
            <input
              id="isbn13"
              placeholder="XXXXXXXXXXXXX"
              type="text"
              {...register("isbn13")}
              className={`
                border ${errors.isbn13 ? 'border-red-500' : 'border-gray-300'}
                bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500
                `}
            />
            {errors.isbn13 && <p className="text-red-500">{errors.isbn13.message}</p>}
          </div>

          {/* Formats */}
          <div className="sm:col-span-2">
            <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Formats<span className="text-red-600 ml-px">*</span></label>
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
              {errors.formats && <p className="text-red-500">Please select at least one book format</p>}
          </div>

          {/* Language */}
          <div className="w-full">
            <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white" htmlFor="language">Language<span className="text-red-600 ml-px">*</span></label>
            <input
              id="language"
              placeholder="en"
              type="text"
              {...register("language")}
              className={`
                border ${errors.language ? 'border-red-500' : 'border-gray-300'}
                className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500"
                `}
            />
            {errors.language && <p className="text-red-500">{errors.language.message}</p>}
          </div>

          {/* Page Count */}
          <div className="w-full">
            <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white" htmlFor="pageCount">Page Count<span className="text-red-600 ml-px">*</span></label>
            <input
              id="pageCount"
              type="number"
              {...register("pageCount", { valueAsNumber: true })}
              className={`
                border ${errors.pageCount ? 'border-red-500' : 'border-gray-300'}
                bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500
                `}
            />
            {errors.pageCount && <p className="text-red-500">{errors.pageCount.message}</p>}
          </div>

          {/* Image Links Field Array */}
          <div className="sm:col-span-2">
            <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Image Links<span className="text-red-600 ml-px">*</span></label>
            <div className="border border-gray-300 rounded">
                {imageLinkFields.map((item, index) => (
                  <div key={item.id} className="flex items-center p-4">
                    <Controller
                      render={({ field }) => <input {...field} className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />}
                      name={`imageLinks.${index}`}
                      control={control}
                    />
                    <button type="button" onClick={() => imageLinkFields.length > 1 && removeImageLink(index)} className="ml-5 bg-dark-clay">
                      <IoClose size={20}/>
                    </button>
                  </div>
                ))}
              <button type="button" onClick={() => appendImageLink('')} className="flex flex-row justify-between items-center bg-dark-clay m-4">
                <IoAddOutline size={20} className="mr-1"/>
                Add Image Link
              </button>
            </div>
            {errors.imageLinks && <p className="text-red-500">{errors.imageLinks.message}</p>}
          </div>


          {/* Description */}
          <div className="sm:col-span-2">
            <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white" htmlFor="description">Description<span className="text-red-600 ml-px">*</span></label>
            <textarea
              id="description"
              rows={4}
              {...register("description")}
              className={`
                border ${errors.description ? 'border-red-500' : 'border-gray-300'}
                block p-2.5 w-full text-sm text-gray-900 bg-gray-50 rounded border border-gray-300 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500
                `}
            />
            {errors.description && <p className="text-red-500">{errors.description.message}</p>}
          </div>

          {/* Notes */}
          <div className="sm:col-span-2">
            <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white" htmlFor="notes">Notes<span className="text-red-600 ml-px">*</span></label>
            <textarea
              id="notes"
              rows={4}
              {...register("notes")}
              className={`
                border ${errors.notes ? 'border-red-500' : 'border-gray-300'}
                block p-2.5 w-full text-sm text-gray-900 bg-gray-50 rounded border border-gray-300 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500
                `}
            />
            {errors.notes && <p className="text-red-500">{errors.notes.message}</p>}
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