import { useEffect, useState, useMemo, useCallback } from 'react';
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
import Delta from 'quill-delta';
import QuillEditor from '../Quill/QuillEditor';
import { transformBookData } from '../../utils/bookFormHelpers';
import 'quill/dist/quill.snow.css';

interface BookFormProps {
  initialData?: Book;
  onSubmit: SubmitHandler<BookFormData>
  isEditMode?: boolean;
  onDelete?: () => void;
  renderAISummaryBtn?: React.ReactNode;
  isLoading?: boolean;
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
  const bookDataEmpty = useMemo(() => _.isEmpty(initialData), [initialData]);

  const parseQuillContent = (content: any): Delta => {
    try {
      if (!content) return new Delta();

      // If content is already a Delta instance
      if (content instanceof Delta) return content;

      // If content is a stringified Delta
      if (typeof content === 'string') {
        const parsed = JSON.parse(content);
        if (parsed && Array.isArray(parsed.ops)) {
          return new Delta(parsed.ops);
        }
      }

      // If content is a Delta-like object
      if (content && typeof content === 'object' && Array.isArray(content.ops)) {
        return new Delta(content.ops);
      }

      // Fallback: create simple Delta with the content
      return new Delta().insert(String(content));
    } catch (error) {
      console.error('Error parsing Quill content:', error);
      return new Delta();
    }
  };

  const initialDescription = useMemo(() =>
    parseQuillContent(initialData?.description), [initialData?.description]
  );

  const initialNotes = useMemo(() =>
    parseQuillContent(initialData?.notes), [initialData?.notes]
  );

  // Original version
  // const initialDescription = useMemo(() => {
  //   if (initialData?.description) {
  //     return typeof initialData.description === 'string'
  //       ? new Delta(JSON.parse(initialData.description))
  //       : new Delta(initialData.description);
  //   }
  //   return new Delta();
  // }, [initialData?.description]);

  // const initialNotes = useMemo(() => {
  //   if (initialData?.notes) {
  //     return typeof initialData.notes === 'string'
  //       ? new Delta(JSON.parse(initialData.notes))
  //       : new Delta(initialData.notes);
  //   }
  //   return new Delta();
  // }, [initialData?.notes]);

  const [description, setDescription] = useState<Delta>(initialDescription);
  const [notes, setNotes] = useState<Delta>(initialNotes);

  const defaultValues = useMemo(() => transformBookData(initialData || {}, formattedDate), [initialData, formattedDate]);

  const {
    control,
    handleSubmit,
    register,
    reset,
    watch,
    formState: { errors },
  } = useForm<BookFormData>({
    resolver: zodResolver(bookSchema),
    defaultValues,
    shouldUseNativeValidation: true,
  });

  const { fields: authorFields, append: appendAuthor, remove: removeAuthor } = useFieldArray({
    control,
    name: "authors",
  });

  const { fields: genreFields, append: appendGenre, remove: removeGenre } = useFieldArray({
    control,
    name: "genres",
  });

  const { fields: tagFields, append: appendTag, remove: removeTag } = useFieldArray({
    control,
    name: "tags",
  });

  // Handle form reset when initialData changes
  useEffect(() => {
    if (initialData) {
      reset(transformBookData(initialData, formattedDate));
    }
  }, [initialData, reset, formattedDate]);

  const onSubmitHandler = useCallback((data: BookFormData) => {
    onSubmit(data);
  }, [onSubmit]);

  const handleDelete = useCallback(() => {
    if (onDelete) {
      onDelete();
    }
  }, [onDelete]);


  // Debugging: console log form values and errors
  const watchedFields = watch();
  useEffect(() => {
    console.log('Current form values:', {
      ...watchedFields,
      description,
      notes
    });
  }, [watchedFields, description, notes]);

  useEffect(() => {
    if (Object.keys(errors).length > 0) {
      console.log("Form errors:", errors);
    }
  }, [errors]);
  return(
    <PageWithErrorBoundary fallbackMessage="Error loading add manual page">
      <form
        className="grid gap-4 grid-cols-2 sm:gap-6"
        onSubmit={handleSubmit(
          onSubmitHandler,
          (errors) => {
            console.log("Form validation errors: ", errors);
          }
        )}
      >

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
                <QuillEditor
                value={field.value instanceof Delta ? field.value : new Delta(field.value)}
                  onChange={(newContent) => {
                    field.onChange(newContent);
                    setDescription(newContent);
                  }}
                  placeholder="Enter book description..."
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
              <QuillEditor
              value={
                field.value instanceof Delta
                  ? field.value
                  : field.value
                  ? new Delta(field.value)
                  : new Delta()
              }
                onChange={(newContent) => {
                  field.onChange(newContent);
                  setNotes(newContent);
                }}
                placeholder="Enter notes..."
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
            onClick={handleDelete}
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
