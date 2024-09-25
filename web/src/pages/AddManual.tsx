import { useEffect } from 'react';
import { Controller, SubmitHandler, useForm, useFieldArray } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useLocation } from 'react-router-dom';
import PageWithErrorBoundary from '../components/ErrorMessages/PageWithErrorBoundary';
import { useFormatPublishDate } from '../utils/formatPublishDate';
import useAddBook from '../hooks/useAddBook';
import _ from 'lodash';
import { languages } from '../consts/languages';
import { TAILWIND_FORM_CLASSES } from '../consts/styleConsts';

import { Book } from '../types/api';
import { IoClose } from 'react-icons/io5';
import { IoAddOutline } from 'react-icons/io5';

const bookSchema = z.object({
  title: z.string().min(1, 'Please enter a title'),
  subtitle: z.string().optional(),
  authors: z.array(z.string().min(1, 'Author name cannot be empty')).min(1, 'Please enter at least one author'),
  genres: z.array(z.string().min(1, 'Genre cannot be empty')).min(1, 'Please enter at least one genre'),
  tags: z.array(z.string().min(1, 'Tag cannot be empty')).min(1, 'At least one tag is required'),
  publishDate: z.string().min(1, 'Please enter a date of publication'),
  isbn10: z.string().length(10, 'ISBN-10 must be 10 characters').optional(),
  isbn13: z.string().length(13, 'ISBN-13 must be 13 characters').optional(),
  formats: z.array(z.enum(['physical', 'eBook', 'audioBook'])).min(1, 'Please select at least one format'),
  language: z.string().min(1, 'Please enter a language'),
  pageCount: z.number().min(1, 'Please enter a total page count'),
  imageLink: z.string().min(1, 'Please enter an image link'),
  description: z.string().min(1, 'Please enter a description'),
  notes: z.string().optional(),
}).refine(
  (data) => data.isbn10 || data.isbn13,
  {
    message: "Either ISBN-10 or ISBN-13 is required",
    path: ["isbn10", "isbn13"],
  }
);

type BookFormData = z.infer<typeof bookSchema>;

const ManualAdd = () => {
  const { mutate: addBook } = useAddBook();
  const location = useLocation();
  const bookData = location.state?.book || {};
  const { formattedDate, dateWarning } = useFormatPublishDate(bookData.publishDate);
  const bookDataEmpty = _.isEmpty(bookData);

  // RH + Zod handlers
  const {
    control,
    handleSubmit,
    register,
    reset,
    setValue,
    setError,
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
    const defaultDate = new Date().toISOString();
    const filteredData = {
      ...data,
      authors: data.authors.filter(author => author.trim() !== ''),
      genres: data.genres.filter(genre => genre.trim() !== ''),
      tags: data.tags.filter(tag => tag.trim() !== ''),
    };
    console.log('Filtered data:', filteredData);
    const validationResult = bookSchema.safeParse(filteredData);
    console.log('Validation result:', validationResult);

    if (!validationResult.success) {
      console.log('Validation failed. Errors:', validationResult.error);
      validationResult.error.issues.forEach(issue => {
        console.log(`Setting error for ${issue.path.join('.')}: ${issue.message}`);
        setError(issue.path.join('.') as any, {
          type: 'manual',
          message: issue.message
        });
      });
      return;
    }

    console.log('Validation successful. Proceeding with book creation.');

    const book: Book = {
      ...data,
      subtitle: data.subtitle || '',
      createdAt: defaultDate,
      lastUpdated: defaultDate,
    };

    addBook(book);
  };


  return (
    <PageWithErrorBoundary fallbackMessage="Error loading add manual page">
      <section className="addManual bg-white min-h-screen bg-cover relative flex flex-col items-center place-content-around px-5 antialiased mdTablet:pr-5 mdTablet:ml-24 dark:bg-black">
        <div className="text-left text-dark max-w-screen-mdTablet pb-24 md:pb-4 flex flex-col relative w-full">
          <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">Add Book</h2>
          <form className="grid gap-4 grid-cols-2 sm:gap-6" onSubmit={handleSubmit(onSubmit)}>

            {/* Title */}
            <div className={TAILWIND_FORM_CLASSES['TWO_COL_WRAPPER']}>
              <label className={TAILWIND_FORM_CLASSES['LABEL']} htmlFor="title">
                Title<span className={TAILWIND_FORM_CLASSES['LABEL_ASTERISK']}>*</span>
              </label>
              <input
                className={`${TAILWIND_FORM_CLASSES['INPUT']} ${errors.title ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
                id="title"
                {...register('title')}
              />
              {errors.title && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.title.message}</p>}
            </div>

            {/* Subtitle */}
            <div className={TAILWIND_FORM_CLASSES['TWO_COL_WRAPPER']}>
              <label
                className={TAILWIND_FORM_CLASSES['LABEL']}
                htmlFor="subtitle"
              >
                  Subtitle (if applicable)
              </label>
              <input
                className={`${TAILWIND_FORM_CLASSES['INPUT']} ${errors.subtitle ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
                id="subtitle" {...register('subtitle')}
              />
              {errors.subtitle && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.subtitle.message}</p>}
            </div>

            {/* Authors Field Array */}
            <div className={TAILWIND_FORM_CLASSES['ONE_COL_WRAPPER']}>
              <label className={TAILWIND_FORM_CLASSES['LABEL']}>
                Authors<span className="text-red-600 ml-px">*</span>
              </label>
              <div className={`${TAILWIND_FORM_CLASSES['FIELD_ARR_WRAPPER']} ${errors.authors ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}>
                <div className={TAILWIND_FORM_CLASSES['FIELD_ARR_COL_WRAPPER']}>
                  {authorFields.map((item, index) => (
                    <div className={TAILWIND_FORM_CLASSES['FIELD_ARR_COL_WRAPPER']} key={item.id}>
                      <Controller
                        render={({ field }) => (
                          <div className={TAILWIND_FORM_CLASSES['FIELD_ARR_ROW_WRAPPER']}>
                            <input
                              {...field}
                              className={`${TAILWIND_FORM_CLASSES['INPUT']} ${errors.authors?.[index] ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''}`}
                            />
                            <button
                              type="button"
                              onClick={() => authorFields.length > 1 && removeAuthor(index)}
                              className={TAILWIND_FORM_CLASSES['REMOVE_BUTTON']}
                            >
                              <IoClose size={20}/>
                            </button>
                          </div>
                        )}
                        name={`authors.${index}`}
                        control={control}
                      />
                      {errors.authors?.[index] && (
                        <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.authors[index]?.message}</p>
                      )}
                    </div>
                  ))}
                </div>
                <button
                  type="button"
                  onClick={() => appendAuthor('')}
                  className={TAILWIND_FORM_CLASSES['ADD_BUTTON']}
                >
                  <IoAddOutline size={20} className="mr-1"/>
                  Add Author
                </button>
              </div>
              {errors.authors && !Array.isArray(errors.authors) && (
                <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.authors.message}</p>
              )}
            </div>

            {/* Genres Field Array */}
            <div className={TAILWIND_FORM_CLASSES['ONE_COL_WRAPPER']}>
              <label className={TAILWIND_FORM_CLASSES['LABEL']}>
                Genres<span className="text-red-600 ml-px">*</span>
              </label>
              <div className={`${TAILWIND_FORM_CLASSES['FIELD_ARR_WRAPPER']} ${errors.genres ? TAILWIND_FORM_CLASSES['ERROR_BORDER']: ''} `}>
                <div className={TAILWIND_FORM_CLASSES['FIELD_ARR_COL_WRAPPER']}>
                  {genreFields.map((item, index) => (
                    <div
                      key={item.id}
                      className={TAILWIND_FORM_CLASSES['FIELD_ARR_COL_WRAPPER']}
                    >
                      <Controller
                        render={({ field }) => (
                          <div className={TAILWIND_FORM_CLASSES['FIELD_ARR_ROW_WRAPPER']}>
                            <input
                              {...field}
                              className={`${TAILWIND_FORM_CLASSES['INPUT']} ${errors.genres?.[index] ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''}`}
                            />
                            <button
                              type="button"
                              onClick={() => genreFields.length > 1 && removeGenre(index)}
                              className={TAILWIND_FORM_CLASSES['REMOVE_BUTTON']}
                            >
                              <IoClose size={20}/>
                            </button>
                          </div>
                        )}
                        name={`genres.${index}`}
                        control={control}
                      />
                      {errors.genres?.[index] && (
                        <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.genres[index]?.message}</p>
                      )}
                    </div>
                  ))}
                </div>
                <button
                  type="button"
                  onClick={() => appendGenre('')}
                  className={TAILWIND_FORM_CLASSES['ADD_BUTTON']}
                >
                  <IoAddOutline size={20} className="mr-1"/>
                  Add Genre
                </button>
              </div>
              {errors.genres && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.genres.message}</p>}
            </div>

            {/* Tags Field Array */}
            <div className={TAILWIND_FORM_CLASSES['ONE_COL_WRAPPER']}>
              <label className={TAILWIND_FORM_CLASSES['LABEL']}>
                Personal Tags<span className={TAILWIND_FORM_CLASSES['LABEL_ASTERISK']}>*</span>
              </label>
              <div className={`${TAILWIND_FORM_CLASSES['FIELD_ARR_WRAPPER']} ${errors.tags ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}>
                <div className={TAILWIND_FORM_CLASSES['FIELD_ARR_COL_WRAPPER']}>
                  {tagFields.map((item, index) => (
                    <div key={item.id} className={TAILWIND_FORM_CLASSES['FIELD_ARR_COL_WRAPPER']}>
                        <Controller
                          render={({ field }) => (
                            <div className={TAILWIND_FORM_CLASSES['FIELD_ARR_ROW_WRAPPER']}>
                              <input
                                {...field}
                                className={`${TAILWIND_FORM_CLASSES['INPUT']} ${errors.tags?.[index] ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''}`}
                              />
                              <button
                                type="button"
                                onClick={() => tagFields.length > 1 && removeTag(index)}
                                className={TAILWIND_FORM_CLASSES['REMOVE_BUTTON']}
                              >
                                <IoClose size={20}/>
                              </button>
                            </div>
                          )}
                          name={`tags.${index}`}
                          control={control}
                        />
                      {errors.tags?.[index] && (
                        <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.tags[index]?.message}</p>
                      )}
                      </div>
                    ))}
                </div>
                <button
                  type="button"
                  onClick={() => appendTag('')}
                  className={TAILWIND_FORM_CLASSES['ADD_BUTTON']}
                >
                  <IoAddOutline size={20} className="mr-1"/>
                  Add Tag
                </button>
              </div>
              {errors.tags && <p className={TAILWIND_FORM_CLASSES['ERROR']}>Please enter at least one personal tag</p>}
            </div>

            {/* Publish Date */}
            <div className={TAILWIND_FORM_CLASSES['ONE_COL_WRAPPER']}>
              <label htmlFor="publishDate" className={TAILWIND_FORM_CLASSES['LABEL']}>
                Publish Date<span className={TAILWIND_FORM_CLASSES['LABEL_ASTERISK']}>*</span>
              </label>
              <input
                type="date"
                id="publishDate"
                min="1000-01-01" {...register('publishDate')}
                className={`${TAILWIND_FORM_CLASSES['INPUT']} ${errors.publishDate ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
              />
              {errors.publishDate && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.publishDate.message}</p>}
              {!bookDataEmpty && dateWarning && <p className="text-yellow-500">{dateWarning}</p>}
            </div>

            {/* ISBN10 */}
            <div className={TAILWIND_FORM_CLASSES['ONE_COL_WRAPPER']}>
              <label htmlFor="isbn10" className={TAILWIND_FORM_CLASSES['LABEL']}>
                ISBN-10<span className={TAILWIND_FORM_CLASSES['LABEL_ASTERISK']}>*</span>
              </label>
              <input
                id="isbn10" {...register('isbn10')}
                className={`${TAILWIND_FORM_CLASSES['INPUT']} ${errors.isbn10 ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
              />
              {errors.isbn10 && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.isbn10.message}</p>}
            </div>

            {/* ISBN13 */}
            <div className={TAILWIND_FORM_CLASSES['ONE_COL_WRAPPER']}>
              <label htmlFor="isbn13" className={TAILWIND_FORM_CLASSES['LABEL']}>
                ISBN-13<span className={TAILWIND_FORM_CLASSES['LABEL_ASTERISK']}>*</span>
              </label>
              <input
                id="isbn13" {...register('isbn13')}
                className={`${TAILWIND_FORM_CLASSES['INPUT']} ${errors.isbn13 ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
              />
              {errors.isbn13 && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.isbn13.message}</p>}
            </div>

            {/* Formats */}
            <div className={TAILWIND_FORM_CLASSES['ONE_COL_WRAPPER']}>
              <label
                htmlFor="checkbox"
                className={TAILWIND_FORM_CLASSES['LABEL']}>
                Formats (select all that apply)<span className={TAILWIND_FORM_CLASSES['LABEL_ASTERISK']}>*</span>
              </label>
              <ul className="grid w-full gap-6 lgMobile:grid-cols-3 mb-1">
                {['physical', 'eBook', 'audioBook'].map((format) => (
                  <li key={format}>
                    <input
                      className="hidden peer"
                      id={`formats_${format}`}
                      type="checkbox"
                      value={format}
                      {...register('formats')}
                    />
                    <label
                      htmlFor={`formats_${format}`}
                      className="inline-flex text-center items-center transition duration-500 shadow-sm justify-center w-full p-2 text-charcoal bg-white border-2 border-gray-200 rounded-lg cursor-pointer dark:hover:text-gray-300 dark:border-gray-700 peer-checked:border-lime-500  hover:bg-gray-200 dark:peer-checked:text-gray-300 peer-checked:text-gray-600 dark:text-white dark:bg-gray-800 dark:hover:bg-gray-700"
                    >
                      {format}
                    </label>
                  </li>
                ))}
              </ul>
              {errors.formats && <p className={TAILWIND_FORM_CLASSES['ERROR']}>Please select at least one book format</p>}
            </div>

            {/* Language */}
            <div className={TAILWIND_FORM_CLASSES['ONE_COL_WRAPPER']}>
              <label
                htmlFor="language"
                className={TAILWIND_FORM_CLASSES['LABEL']}>
                Language <span className={TAILWIND_FORM_CLASSES['LABEL_ASTERISK']}>*</span>
              </label>
              <Controller
                name="language"
                control={control}
                render={({ field }) => (
                  <select
                    {...field}
                    className={`${TAILWIND_FORM_CLASSES['TEXT_AREA']} ${errors.language ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
                  >
                    {languages.map((language) => (
                      <option key={language.value} value={language.value}>
                        {language.label}
                      </option>
                    ))}
                  </select>
                )}
              />
                {errors.language && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.language.message}</p>}
            </div>

            {/* Page Count */}
            <div className={TAILWIND_FORM_CLASSES['ONE_COL_WRAPPER']}>
              <label
                className={TAILWIND_FORM_CLASSES['LABEL']}
                htmlFor="pageCount"
              >
                Page Count<span className={TAILWIND_FORM_CLASSES['LABEL_ASTERISK']}>*</span>
              </label>
              <input
                id="pageCount"
                type="number"
                {...register('pageCount', { valueAsNumber: true })}
                className={`${TAILWIND_FORM_CLASSES['INPUT']} ${errors.pageCount ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
              />
              {errors.pageCount && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.pageCount.message}</p>}
            </div>

            {/* Image Link */}
            <div className={TAILWIND_FORM_CLASSES['ONE_COL_WRAPPER']}>
              <label
                className={TAILWIND_FORM_CLASSES['LABEL']}
                htmlFor="imageLink"
              >
                Image Link<span className={TAILWIND_FORM_CLASSES['LABEL_ASTERISK']}>*</span>
              </label>
              <input
                type="url"
                id="imageLink"
                className={`${TAILWIND_FORM_CLASSES['INPUT']} ${errors.imageLink ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
                {...register('imageLink')}
              />
              {errors.imageLink && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.imageLink.message}</p>}
            </div>

            {/* Description */}
            <div className={TAILWIND_FORM_CLASSES['ONE_COL_WRAPPER']}>
              <label
                htmlFor="description"
                className={TAILWIND_FORM_CLASSES['LABEL']}
              >
                Description<span className={TAILWIND_FORM_CLASSES['LABEL_ASTERISK']}>*</span>
              </label>
              <textarea
                className={`${TAILWIND_FORM_CLASSES['TEXT_AREA']} ${errors.description ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
                id="description"
                rows={4}
                {...register('description')}
              />
              {errors.description && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.description.message}</p>}
            </div>

            {/* Notes */}
            <div className={TAILWIND_FORM_CLASSES['ONE_COL_WRAPPER']}>
              <label
                htmlFor="notes"
                className={TAILWIND_FORM_CLASSES['LABEL']}
              >
                Notes (optional)
              </label>
              <textarea
                className={TAILWIND_FORM_CLASSES['TEXT_AREA']}
                id="notes"
                rows={4}
                {...register('notes')}
              />
              {errors.notes && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.notes.message}</p>}
            </div>

            <button
              className="bg-vivid-blue hover:bg-vivid-blue-d hover:border-vivid-blue transition duration-300 ease-in-out"
              type="submit"
            >
              Add Book
            </button>
          </form>
        </div>
      </section>
    </PageWithErrorBoundary>
  );
};

export default ManualAdd;
