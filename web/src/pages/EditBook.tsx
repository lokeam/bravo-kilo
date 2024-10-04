import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Controller, useForm, useFieldArray, SubmitHandler } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Markdown from 'react-markdown';
import DOMPurify from 'dompurify';
import PageWithErrorBoundary from '../components/ErrorMessages/PageWithErrorBoundary';

import useUpdateBook from '../hooks/useUpdateBook';
import useDeleteBook from '../hooks/useDeleteBook';
import useFetchBookById from '../hooks/useFetchBookById';
import Loading from '../components/Loading/Loading';
import { languages } from '../consts/languages';
import { TAILWIND_FORM_CLASSES } from '../consts/styleConsts';

import Modal from '../components/Modal/ModalRoot';
import BookSummaryBtn from '../components/BookSummaryBtn/BookSummaryBtn';
import { BookFormData } from '../types/api';

import { IoClose } from 'react-icons/io5';
import { IoAddOutline } from 'react-icons/io5';
import { IoIosWarning } from "react-icons/io";
import { MdDeleteForever } from "react-icons/md";
import { RiFileCopyLine } from "react-icons/ri";


const bookSchema = z.object({
  title: z.string().min(1, 'Please enter a title'),
  subtitle: z.string().optional(),
  authors: z.array(z.string()).min(1, 'Please enter at least one author'),
  genres: z.array(z.string()).min(1, 'Please enter at least one genre'),
  tags: z.array(z.string()).min(1, 'At least one tag is required'),
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

type BookFormData = {
  title: string;
  subtitle?: string;
  authors: { author: string }[]; // React Hook Form needs field array strings saved in an object
  genres: { genre: string }[];
  tags: { tag: string }[];
  publishDate: string;
  isbn10: string;
  isbn13: string;
  formats: ("physical" | "eBook" | "audioBook")[] | undefined;
  language: string;
  pageCount: number;
  imageLink: string;
  description: string;
  notes: string;
}


const EditBook = () => {
  // Delete Modal state
  const [opened, setOpened] = useState(false);

  // AI Preview + Modal state
  const [aiSummaryPreview, setAiSummaryPreview] = useState("");
  const [isPreviewModalOpen, setIsPreviewModalOpen] = useState(false);
  const [isManualTrigger, setIsManualTrigger] = useState(false);
  const { bookID } = useParams();

  console.log('-------------------------');
  console.log('EditBook page');
  console.log('bookID: ', bookID);
  const { data: book, isLoading, isError } = useFetchBookById(bookID as string, !!bookID);
  const { mutate: updateBook } = useUpdateBook(bookID as string);
  const { mutate: deleteBook } = useDeleteBook();

  /* React hook form handlers */
  const {
    register,
    handleSubmit,
    setError,
    control,
    reset,
    formState: { errors }
  } = useForm<BookFormData>({
    resolver: zodResolver(bookSchema),
  });


  const {
    fields: authorFields,
    append: appendAuthor,
    remove: removeAuthor
  } = useFieldArray({
    control,
    name: "authors" as const,
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
    if (book) {
      const publishDateFormattedBook = {
        ...book,
        publishDate: book.publishDate ? new Date(book.publishDate).toISOString().split('T')[0] : '',
        authors: book.authors?.map((author) => ({ author })) || [],
        genres: book.genres?.map((genre) => ({ genre })) || [],
        tags: book.tags?.map((tag) => ({ tag })) || [],
      };
      reset(publishDateFormattedBook);
    }
  }, [book, reset]);


  // Debug useEffect for Preview Modal
  useEffect(() => {
    console.log('isPreviewModalOpen state changed:', isPreviewModalOpen);
  }, [isPreviewModalOpen]);

  if (isLoading) return <Loading />;

  if (isError) return <div>Error loading book data</div>;

  // Form Submittal
  const onSubmit: SubmitHandler<BookFormData> = (data) => {
    console.log(`Form submitted with data ${data}`);
    const defaultDate = new Date().toISOString();
    const filteredData = {
      ...data,
      authors: data.authors.filter(item => item.author.trim() !== '').map(item => item.author),
      genres: data.genres.filter(item => item.genre.trim() !== '').map(item => item.genre),
      tags: data.tags.filter(item => item.tag.trim() !== '').map(item => item.tag),
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

    const book = {
      ...data,
      id: Number(bookID),
      createdAt: defaultDate,
      lastUpdated: defaultDate,
      isbn10: data.isbn10 || '',
      isbn13: data.isbn13 || '',
      authors: data.authors.map((authorObj) => authorObj.author),
      genres: data.genres.map((genreObj) => genreObj.genre),
      tags: data.tags.map((tagObj) => tagObj.tag),
    };

    updateBook(book);
  };

  const copyTextContent = () => {
    const element = document.getElementById("prompt_response");

    if (element) {
      // Get the text content with appropriate handling for line breaks
      const textContent = Array.from(element.childNodes)
        .map(node => (node.textContent || '').trim())
        .join('\n\n'); // Use double newline for clearer formatting

      // Copy the content using the Clipboard API
      navigator.clipboard.writeText(textContent).then(
        () => alert("Text copied to clipboard!"),
        (err) => console.error("Failed to copy text: ", err)
      );
    }
  };

  // Delete Modal
  const openModal = () => setOpened(true);
  const closeModal = () => setOpened(false);
  const handleDelete = (event: React.MouseEvent<HTMLButtonElement>) => {
    event?.preventDefault;
    deleteBook(bookID as string)
  };

  // Preview Modal
  const openPreviewModal = () => {
    console.log('openPreviewModal fired');

    setIsManualTrigger(true);
    setIsPreviewModalOpen(true)
  };

  const closePreviewModal = () => {
    console.log('closePreviewModal fired');

    setIsManualTrigger(false);
    setIsPreviewModalOpen(false);

    console.log('state check within closePreviewModal handler: setIsPreviewModalOpen: ', isPreviewModalOpen);
  };

  console.log('RHF Errors: ', errors);
  console.log('state check OUTSIDE of closePreviewModal handler: setIsPreviewModalOpen: ', isPreviewModalOpen);

  const promptTitle = book?.title || "";
  const promptAuthors = book?.authors || [""];

  console.log('Testing book data from useFetchBookById: ', book);

  return (
    <PageWithErrorBoundary fallbackMessage="Error loading edit page">
      <section className="editBook bg-white min-h-screen bg-cover relative flex flex-col items-center place-content-around px-5 antialiased mdTablet:pr-5 mdTablet:ml-24 dark:bg-black">
        <div className="text-left text-dark max-w-screen-mdTablet pb-24 md:pb-4 flex flex-col relative w-full">
          <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">Edit Book</h2>
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
                      {errors.authors?.[index] && (
                        <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.authors[index]?.message}</p>
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
                      {errors.genres?.[index] && (
                        <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.genres[index]?.message}</p>
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
                      {errors.tags?.[index] && (
                        <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.tags[index]?.message}</p>
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
                className={`${TAILWIND_FORM_CLASSES['INPUT']} ${errors.publishDate ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
              />
              {errors.publishDate && <p className={TAILWIND_FORM_CLASSES['ERROR']}>{errors.publishDate.message}</p>}
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
              <div className="border border-cadet-gray rounded p-4">
              <textarea
                className={`${TAILWIND_FORM_CLASSES['TEXT_AREA']} ${errors.description ? TAILWIND_FORM_CLASSES['ERROR_BORDER'] : ''} `}
                id="description"
                rows={4}
                {...register('description')}
              />
                <div className="grid w-full gap-6 lgMobile:grid-cols-3 pt-2">
                  <BookSummaryBtn
                    title={promptTitle}
                    authors={promptAuthors}
                    setAiSummaryPreview={setAiSummaryPreview}
                    openPreviewModal={openPreviewModal}
                    isPreviewModalOpen={isPreviewModalOpen}
                    isManualTrigger={isManualTrigger}
                    setIsManualTrigger={setIsManualTrigger}
                  />
                </div>
              </div>

              {errors.description && <p className="text-red-500">{errors.description.message}</p>}
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
              Update Book
            </button>
            <button
              type="button"
              onClick={openModal}
              className="border-red-500 text-red-500 hover:text-white hover:bg-red-600 focus:ring-red-900"
            >
              Delete Book
            </button>
          </form>

          {/* Delete Modal */}
          <Modal opened={opened} onClose={closeModal}>
            <div className="flex items-center justify-center">
              <IoIosWarning size={30} />
            </div>
            <h3 className="flex items-center justify-center text-lg">Are you sure that you want to delete this book?</h3>
            <p className="flex items-center justify-center mb-5">This action cannot be undone.</p>
            <button type="button" onClick={closeModal} className="flex flex-row justify-between items-center bg-transparent mr-1 w-full mb-3 lg:mb-0">
              <span>Cancel</span>
            </button>
            <button type="button" onClick={handleDelete} className="flex flex-row justify-between items-center bg-transparent mr-1 w-full text-white bg-red-600 hover:bg-red-800 focus:ring-red-800 mb-3 lg:mb-0">
              <span>Yes, I want to delete this book</span>
              <MdDeleteForever size={30}/>
            </button>
          </Modal>

          {/* Gemini Modal */}
          {isPreviewModalOpen && (
            <Modal opened={isPreviewModalOpen} onClose={closePreviewModal}>
              <div className="summary_modal p-4">
                <h3 className="mb-2 text-2xl font-semibold text-indigo-900 dark:text-white">AI-Generated Summary Preview</h3>
                <div id="prompt_response" className="prompt_response break-words">
                  <Markdown>{DOMPurify.sanitize(aiSummaryPreview)}</Markdown>
                </div>
                <div className="flex justify-end space-x-4">
                  <button
                    className="flex flex-row justify-between bg-transparent border border-gray-600"
                    onClick={copyTextContent}>
                      <RiFileCopyLine size={22} className="pt-1 mr-2" color="white"/>
                    <span>Copy text</span>
                  </button>
                  <button onClick={() => {
                    console.log('Summary Modal close button clicked');
                    closePreviewModal();
                  }}>Close Modal</button>
                </div>
              </div>
            </Modal>
          )}
        </div>
      </section>
    </PageWithErrorBoundary>

  );
};

export default EditBook;
