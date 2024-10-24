import { useEffect } from 'react';
import { Controller, SubmitHandler, useForm, useFieldArray } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';

import PageWithErrorBoundary from '../ErrorMessages/PageWithErrorBoundary';
import { TAILWIND_FORM_CLASSES } from '../../consts/styleConsts';
import { languages } from '../../consts/languages';
import { BookFormData, Book } from '../../types/api';
import { bookSchema } from '../../utils/bookSchema';
import { IoClose, IoAddOutline } from 'react-icons/io5';
import { useFormatPublishDate } from '../../utils/formatPublishDate';
import _ from 'lodash';
import DOMPurify from 'dompurify';
import ReactQuill from 'react-quill';
import Delta from 'quill-delta';
import 'react-quill/dist/quill.snow.css';


interface BookFormProps {
  initialData?: Book;
  onSubmit: (data: StringifiedBookFormData) => void;
  isEditMode?: boolean;
  onDelete?: () => void;
  renderAISummaryBtn?: React.ReactNode;
  isLoading?: boolean;
}

type StringifiedBookFormData = Omit<BookFormData, 'description' | 'notes'> & {
  description: string;
  notes: string | null;
};

function transformBookData(bookData: Partial<Book> = {}, formattedDate: string): BookFormData {
  const parseField = (field: string | undefined | null): Delta => {
    if (!field) return new Delta();
    try {
      // First, try to parse as JSON
      const parsed = JSON.parse(field);
      return new Delta(Array.isArray(parsed.ops) ? parsed : { ops: [{ insert: field }] });
    } catch (error) {
      // If parsing fails, assume it's plain text and create a Delta object
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
}

function BookForm({
  initialData,
  onSubmit,
  isEditMode = false,
  onDelete,
  renderAISummaryBtn,
  isLoading = false,
}: BookFormProps) {
  const { formattedDate, dateWarning } = useFormatPublishDate(initialData?.publishDate || '');
  const bookDataEmpty = _.isEmpty(initialData);

  const {
    control,
    handleSubmit,
    register,
    reset,
    watch, // Watch the form data for debugging
    formState: { errors },
  } = useForm<BookFormData>({
    resolver: zodResolver(bookSchema),
    defaultValues: transformBookData(initialData || {}, formattedDate),
    shouldUseNativeValidation: true,
  });

  // Init field arrays
  const {
    fields: authorFields,
    append: appendAuthor,
    remove: removeAuthor,
  } = useFieldArray({
    control,
    name: 'authors' as const,
  });

  const {
    fields: genreFields,
    append: appendGenre,
    remove: removeGenre,
  } = useFieldArray({
    control,
    name: 'genres' as const,
  });

  const {
    fields: tagFields,
    append: appendTag,
    remove: removeTag,
  } = useFieldArray({
    control,
    name: 'tags' as const,
  });

  // Watch the form data for debugging
  const watchedFields = watch();
  useEffect(() => {
    console.log('Current form values:', watchedFields);
  }, [watchedFields]);


  // Handle form reset when initialData changes
  useEffect(() => {
    if (initialData) {
      reset(transformBookData(initialData, formattedDate));
    }
  }, [initialData, reset, formattedDate]);


  const onSubmitWithJSONB: SubmitHandler<BookFormData> = (data) => {
    const jsonbData: StringifiedBookFormData = {
      ...data,
      description: JSON.stringify(data.description),
      notes: data.notes ? JSON.stringify(data.notes) : null,
    };

    // Console log the form data
    console.log('Form Data:', data);
    console.log('Stringified JSONB Data:', jsonbData);

    onSubmit(jsonbData);
  };

  return(
    <PageWithErrorBoundary fallbackMessage="Error loading add manual page">
      <form className="grid gap-4 grid-cols-2 sm:gap-6" onSubmit={handleSubmit(onSubmitWithJSONB)}>

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
                    name={`authors.${index}.author`}
                    control={control}
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
                  />
                  {errors.authors?.[index]?.author && (
                    <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.authors[index]?.author?.message}</p>
                  )}
                </div>
              ))}
            </div>
            <button
              type="button"
              onClick={() => appendAuthor({author: ''})}
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
                    name={`genres.${index}.genre`}
                    control={control}
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

                  />
                  {
                  errors.genres?.[index] && (
                    <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.genres[index]?.genre?.message}</p>
                  )}
                </div>
              ))}
            </div>
            <button
              type="button"
              onClick={() => appendGenre({genre: ''})}
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
                      name={`tags.${index}.tag`}
                      control={control}
                      rules={{ required: 'Tag field cannot be empty '}}
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
                    />
                  {errors.tags?.[index]?.tag && (
                    <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.tags[index]?.tag?.message}</p>
                  )}
                  </div>
                ))}
            </div>
            <button
              type="button"
              onClick={() => appendTag({tag: ''})}
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
            className={`custom_date_input ${TAILWIND_FORM_CLASSES['INPUT']} ${errors.publishDate ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
          />
          {errors.publishDate && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.publishDate.message}</p>}
          {!bookDataEmpty && dateWarning && <p className="text-orange-500 dark:text-yellow-500">{dateWarning}</p>}
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
                  className="inline-flex text-center items-center transition duration-500 shadow-sm justify-center w-full p-2 text-charcoal bg-white border-2 border-gray-200 rounded-lg cursor-pointer dark:hover:text-gray-300 dark:border-gray-700/60 peer-checked:border-lime-500  hover:bg-gray-200 dark:peer-checked:text-gray-300 peer-checked:text-gray-600 dark:text-white dark:bg-gray-800 dark:hover:bg-gray-700"
                >
                  {format}
                </label>
              </li>
            ))}
          </ul>
          {errors.formats && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.formats.message}</p>}
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
                className={`${TAILWIND_FORM_CLASSES['INPUT']} ${errors.language ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
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
          <div className={TAILWIND_FORM_CLASSES['FIELD_ARR_WRAPPER']}>
            <Controller
              name="description"
              control={control}
              render={({ field }) => (
                <ReactQuill
                theme="snow"
                value={field.value && field.value.ops ? JSON.stringify(field.value) : ''}
                // @ts-ignore
                onChange={(content, delta, source, editor) => {
                  field.onChange(editor.getContents());
                }}
                className={`${TAILWIND_FORM_CLASSES['INPUT']} ${errors.description ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
              />
              )}
            />
            {/* Render AI Summary Btn if EditBook page */}
            {renderAISummaryBtn}
          </div>
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
          <Controller
            name="notes"
            control={control}
            render={({ field }) => (
              <ReactQuill
              theme="snow"
              value={field.value && field.value.ops ? JSON.stringify(field.value) : ''}
              // @ts-ignore
              onChange={(content, delta, source, editor) => {
                field.onChange(editor.getContents());
              }}
              className={`${TAILWIND_FORM_CLASSES['INPUT']} ${errors.notes ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
            />
            )}
          />
          {errors.notes && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.notes.message}</p>}
        </div>

        {/* Submit Button */}
        <button
          className="bg-vivid-blue hover:bg-vivid-blue-d hover:border-vivid-blue-d dark:bg-vivid-blue dark:hover:bg-vivid-blue-d dark:hover:border-vivid-blue-d dark:hover:text-white transition duration-500 ease-in-out"
          type="submit"
          disabled={isLoading}
        >
          {isLoading ? 'Loading...' : isEditMode ? 'Update Book' : 'Add Book' }
        </button>

        {/* Delete Button */}
        {isEditMode && onDelete && (
          <button
            type="button"
            onClick={onDelete}
            className="bg-transparent border-red-500 text-red-500 hover:text-white dark:hover:text-white hover:bg-red-800 focus:ring-red-800 hover:border-red-800 dark:hover:bg-red-800 dark:hover:border-red-800 transition duration-500 ease-in-out"
            disabled={isLoading}
          >
            Delete Book
          </button>
        )}

      </form>
    </PageWithErrorBoundary>
  )
}

export default BookForm;
