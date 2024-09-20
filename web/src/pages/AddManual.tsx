import { useEffect } from 'react';
import { Controller, SubmitHandler, useForm, useFieldArray } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useLocation } from 'react-router-dom';
import PageWithErrorBoundary from '../components/ErrorMessages/PageWithErrorBoundary';
import { useFormatPublishDate } from '../utils/formatPublishDate';
import useAddBook from '../hooks/useAddBook';
import LanguageSelect from '../components/LanguageSelect/LanguageSelect';

import { Book } from '../types/api';
import { IoClose } from 'react-icons/io5';
import { IoAddOutline } from 'react-icons/io5';

const bookSchema = z.object({
  title: z.string().min(1, 'Please enter a title'),
  subtitle: z.string().optional(),
  authors: z.array(z.string()).min(1, 'Please enter at least one author'),
  genres: z.array(z.string()).min(1, 'Please enter at least one genre'),
  tags: z.array(z.string()).min(1, 'At least one tag is required'),
  publishDate: z.string().min(1, 'Please enter a date of publication'),
  isbn10: z.string().min(10).max(10, 'This field must contain 10 characters'),
  isbn13: z.string().min(13).max(13,  'This field must contain 13 characters'),
  formats: z.array(z.enum(['physical', 'eBook', 'audioBook'])).min(1, 'Please select at least one format'),
  language: z.string().min(1, 'Please enter a language'),
  pageCount: z.number().min(1, 'Please enter a total page count'),
  imageLink: z.string().min(1, 'Please enter an image link'),
  description: z.string().min(1, 'Please enter a description'),
  notes: z.string().optional(),
});

type BookFormData = z.infer<typeof bookSchema>;

const ManualAdd = () => {
  const { mutate: addBook } = useAddBook();
  const location = useLocation();
  const bookData = location.state?.book || {};
  const { formattedDate, dateWarning } = useFormatPublishDate(bookData.publishDate);

  const {
    control,
    handleSubmit,
    register,
    reset,
    setValue,
    formState: { errors },
  } = useForm<BookFormData>({
    resolver: zodResolver(bookSchema),
    defaultValues: {
      title: '',
      subtitle: '',
      authors: [],
      genres: [],
      tags: [],
      publishDate: formattedDate,
      isbn10: '',
      isbn13: '',
      formats: [],
      language: 'en',
      pageCount: 0,
      imageLink: '',
      description: '',
      notes: '',
      ...bookData,
    },
  });

  useEffect(() => {
    setValue('publishDate', formattedDate);
  }, [formattedDate, setValue]);

  useEffect(() => {
    if (errors) {
      const firstErrorKey = Object.keys(errors)[0];

      if (firstErrorKey) {
        const errorElement = document.querySelector(`[name="${firstErrorKey}]`) as HTMLElement;

        if (errorElement) {
          errorElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
          errorElement.focus();
        }
      }
    }
  }, [errors]);

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


  useEffect(() => {
    // Only reset if bookData is actually different
    if (bookData && Object.keys(bookData).length > 0) {
      reset(bookData);
    }
  }, [bookData, reset]);

  // Update publish date when formattedDate changes
  useEffect(() => {
    setValue('publishDate', formattedDate);
  }, [formattedDate, setValue]);

  const onSubmit: SubmitHandler<BookFormData> = (data) => {
    console.log('Submitted data:', data);
    const defaultDate = new Date().toISOString();

    const book: Book = {
      ...data,
      subtitle: data.subtitle || '',
      createdAt: defaultDate,
      lastUpdated: defaultDate,
    };

    addBook(book);
  };

  console.log('RHF errors: ', errors );

  return (
    <PageWithErrorBoundary fallbackMessage="Error loading add manual page">
      <section className="addManual bg-black relative flex flex-col items-center place-content-around px-5 antialiased mdTablet:pr-5 mdTablet:ml-24 h-screen">
        <div className="text-left max-w-screen-mdTablet py-24 md:pb-4 flex flex-col relative w-full">
          <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">Add Book</h2>
          <form className="grid gap-4 grid-cols-2 sm:gap-6" onSubmit={handleSubmit(onSubmit)}>

            {/* Title */}
            <div className="block col-span-2 mdTablet:col-span-1">
              <label className="block mb-2 text-base font-medium text-gray-900 dark:text-white" htmlFor="title">Title<span className="text-red-600 ml-px">*</span></label>
              <input className={`bg-maastricht border ${errors.title ? 'border-red-500' : 'border-gray-600'} text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500 mb-1`} id="title" {...register('title')} />
              {errors.title && <p className="text-red-500">{errors.title.message}</p>}
            </div>

            {/* Subtitle */}
            <div className="block col-span-2 mdTablet:col-span-1">
              <label className="block mb-2 text-base font-medium text-gray-900 dark:text-white" htmlFor="subtitle">Subtitle (if applicable)</label>
              <input className={`bg-maastricht border ${errors.subtitle ? 'border-red-500' : 'border-gray-600'} text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500 mb-1`} id="subtitle" {...register('subtitle')} />
              {errors.subtitle && <p className="text-red-500">{errors.subtitle.message}</p>}
            </div>

            {/* Authors Field Array */}
            <div className="block col-span-2">
              <label className="block mb-2 text-base font-medium text-gray-900 dark:text-white">Authors<span className="text-red-600 ml-px">*</span></label>
              <div className={`border ${errors.authors ? 'border-red-500' : 'border-cadet-gray'} rounded p-4 mb-1`}>
                {authorFields.map((item, index) => (
                  <div className="flex w-full items-center mb-4 col-span-2" key={item.id}>
                    <Controller
                      render={({ field }) => <input {...field} className="bg-maastricht border border-gray-300 text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5  dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />}
                      name={`authors.${index}`}
                      control={control}
                    />
                    <button type="button" onClick={() => authorFields.length > 1 && removeAuthor(index)}  className="flex flex-row justify-between items-center bg-dark-clay ml-4">
                      <IoClose size={20}/>
                    </button>
                  </div>
                ))}
                <button type="button" onClick={() => appendAuthor('')} className="flex flex-row text-base justify-between items-center bg-dark-clay py-2 px-3">
                  <IoAddOutline size={20} className="mr-1"/>
                  Add Author
                </button>
              </div>
              {errors.authors && <p className="text-red-500">{errors.authors.message}</p>}
            </div>

            {/* Genres Field Array */}
            <div className="block col-span-2">
              <label className="block mb-2 text-base  font-medium text-gray-900 dark:text-white">Genres<span className="text-red-600 ml-px">*</span></label>
              <div className={`border ${errors.genres ? 'border-red-500' : 'border-cadet-gray'} rounded p-4 mb-1`}>
                <div className="flex flex-col sm:gap-6 ">
                {genreFields.map((item, index) => (
                    <div key={item.id} className="flex w-full items-center mb-4 col-span-2">
                      <Controller
                        render={({ field }) => <input {...field} className="bg-maastricht border border-gray-300 text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5  dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500"/>}
                        name={`genres.${index}`}
                        control={control}
                      />
                      <button type="button" onClick={() => genreFields.length > 1 && removeGenre(index)} className="flex flex-row justify-between items-center bg-dark-clay ml-4">
                        <IoClose size={20}/>
                      </button>
                    </div>
                  ))}
                </div>
                <button type="button" onClick={() => appendGenre('')} className="flex flex-row justify-between items-center bg-dark-clay py-2 px-3">
                  <IoAddOutline size={20} className="mr-1"/>
                  Add Genre
                </button>
              </div>
              {errors.genres && <p className="text-red-500">{errors.genres.message}</p>}
            </div>

            {/* Tags Field Array */}
            <div className="block col-span-2">
              <label className="block mb-2 text-base font-medium text-gray-900 dark:text-white">Personal Tags<span className="text-red-600 ml-px">*</span></label>
              <div className={`border ${errors.tags ? 'border-red-500' : 'border-cadet-gray'} rounded p-4 mb-1`}>
                <div className="flex flex-col sm:gap-6 ">
                {tagFields.map((item, index) => (
                    <div key={item.id} className="flex w-full items-center mb-4 col-span-2">
                      <Controller
                        render={({ field }) => <input {...field} className="bg-maastricht border border-gray-300 text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5  dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500"/>}
                        name={`tags.${index}`}
                        control={control}
                      />
                      <button type="button" onClick={() => tagFields.length > 1 && removeTag(index)} className="flex flex-row justify-between items-center bg-dark-clay ml-4">
                        <IoClose size={20}/>
                      </button>
                    </div>
                  ))}
                </div>
                <button type="button" onClick={() => appendTag('')} className="flex flex-row justify-between items-center bg-dark-clay py-2 px-3">
                  <IoAddOutline size={20} className="mr-1"/>
                  Add Tag
                </button>
              </div>
              {errors.tags && <p className="text-red-500">Please enter at least one personal tag</p>}
            </div>

            {/* Publish Date */}
            <div className="block col-span-2">
              <label htmlFor="publishDate" className="block mb-2  text-base  font-medium text-gray-900 dark:text-white">Publish Date<span className="text-red-600 ml-px">*</span></label>
              <input
                type="date"
                id="publishDate"
                min="1000-01-01" {...register('publishDate')}
                className={`bg-maastricht border ${
                  errors.publishDate ? 'border-red-500' : 'border-gray-600'
                } text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500 mb-1`} />
              {errors.publishDate && <p className="text-red-500">{errors.publishDate.message}</p>}
              {dateWarning && <p className="text-yellow-500">{dateWarning}</p>}
            </div>

            {/* ISBN10 */}
            <div className="block col-span-2">
              <label htmlFor="isbn10" className="block mb-2  text-base  font-medium text-gray-900 dark:text-white">ISBN-10<span className="text-red-600 ml-px">*</span></label>
              <input id="isbn10" {...register('isbn10')} className={`bg-maastricht border ${errors.isbn10 ? 'border-red-500' : 'border-gray-600'} text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500 mb-1`} />
              {errors.isbn10 && <p className="text-red-500">{errors.isbn10.message}</p>}
            </div>

            {/* ISBN13 */}
            <div className="block col-span-2">
              <label htmlFor="isbn13" className="block mb-2  text-base  font-medium text-gray-900 dark:text-white">ISBN-13<span className="text-red-600 ml-px">*</span></label>
              <input id="isbn13" {...register('isbn13')} className={`bg-maastricht border ${errors.isbn13 ? 'border-red-500' : 'border-gray-600'} text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500 mb-1`} />
              {errors.isbn13 && <p className="text-red-500">{errors.isbn13.message}</p>}
            </div>

            {/* Formats */}
            <div className="col-span-2">
              <label className="block mb-2 text-base  font-medium text-gray-900 dark:text-white">Formats (select all that apply)<span className="text-red-600 ml-px">*</span></label>
                <ul className="grid w-full gap-6 lgMobile:grid-cols-3 mb-1">
                  {['physical', 'eBook', 'audioBook'].map((format) => (
                    <li key={format}>
                      <input
                        type="checkbox"
                        id={`formats_${format}`}
                        {...register('formats')}
                        value={format}
                        className="hidden peer"
                      />
                      <label htmlFor={`formats_${format}`} className="inline-flex text-center items-center justify-center w-full p-2 text-gray-500 bg-white border-2 border-gray-200 rounded cursor-pointer dark:hover:text-gray-300 dark:border-gray-700 peer-checked:border-margorelle-comp1-g   hover:text-gray-600 dark:peer-checked:text-gray-300 peer-checked:text-gray-600 hover:bg-maastricht dark:text-gray-400 dark:bg-gray-800 dark:hover:bg-gray-700">{format}</label>
                    </li>
                  ))}
                </ul>
                {errors.formats && <p className="text-red-500">Please select at least one book format</p>}
            </div>

            {/* Language */}
            <div className="col-span-2">
              <LanguageSelect control={control} errors={errors} />
            </div>

            {/* Page Count */}
            <div className="col-span-2">
              <label htmlFor="pageCount" className="block mb-2 text-base font-medium text-gray-900 dark:text-white">Page Count<span className="text-red-600 ml-px">*</span></label>
              <input id="pageCount" type="number" {...register('pageCount', { valueAsNumber: true })} className={`bg-maastricht border ${errors.pageCount ? 'border-red-500' : 'border-gray-600'} text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500 mb-1`} />
              {errors.pageCount && <p className="text-red-500">{errors.pageCount.message}</p>}
            </div>

            {/* Image Link */}
            <div className="col-span-2">
              <label className="block mb-2 text-base font-medium text-gray-900 dark:text-white">Image Link<span className="text-red-600 ml-px">*</span></label>
              <input type="url" className={`bg-maastricht border ${errors.imageLink ? 'border-red-500' : 'border-gray-600'} text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500 mb-1`} id="imageLink" {...register('imageLink')} />
              {errors.imageLink && <p className="text-red-500">{errors.imageLink.message}</p>}
            </div>

            {/* Description */}
            <div className="col-span-2">
              <label htmlFor="description" className="block mb-2 text-base font-medium text-gray-900 dark:text-white">Description<span className="text-red-600 ml-px">*</span></label>
              <textarea id="description" rows={4} {...register('description')} className={`block p-2.5 w-full text-base text-gray-900 bg-maastricht rounded border ${errors.description ? 'border-red-500' : 'border-gray-600'} focus:ring-primary-500 focus:border-primary-500 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500 mb-1`} />
              {errors.description && <p className="text-red-500">{errors.description.message}</p>}
            </div>

            {/* Notes */}
            <div className="col-span-2">
              <label htmlFor="notes" className="block mb-2 text-base font-medium text-gray-900 dark:text-white">Notes (optional)</label>
              <textarea id="notes" rows={4} {...register('notes')} className={`block p-2.5 w-full text-base text-gray-900 bg-maastricht rounded border ${errors.notes ? 'border-red-500' : 'border-gray-600'} focus:ring-primary-500 focus:border-primary-500 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500 mb-1`} />
              {errors.notes && <p className="text-red-500">{errors.notes.message}</p>}
            </div>

            <button className="bg-majorelle hover:bg-hepatica" type="submit">
              Add Book
            </button>
          </form>
        </div>
      </section>
    </PageWithErrorBoundary>
  );
};

export default ManualAdd;
