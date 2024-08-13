import { useState, useEffect } from 'react';
import { Controller, useForm, useFieldArray, SubmitHandler } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useNavigate, useParams } from 'react-router-dom';
import useUpdateBook from '../hooks/useUpdateBook';
import useFetchBookById from '../hooks/useFetchBookById';

import Modal from '../components/Modal/ModalRoot';
import { Book } from './Library';

import { IoClose } from 'react-icons/io5';
import { IoAddOutline } from 'react-icons/io5';
import { IoIosWarning } from "react-icons/io";
import { MdDeleteForever } from "react-icons/md";


const bookSchema = z.object({
  id: z.number(),
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

const EditBook = () => {
  const [opened, setOpened] = useState(false);

  const { bookID } = useParams();

  console.log('edit book');
  console.log('bookID: ', bookID);
  const { data: book, isLoading, isError } = useFetchBookById(bookID as string, !!bookID);
  const { mutate: updateBook } = useUpdateBook(bookID as string);
  const navigate = useNavigate();

  /* React hook form handlers */
  const { register, handleSubmit, control, reset, formState: { errors } } = useForm<BookFormData>({
    resolver: zodResolver(bookSchema),
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
      id: Number(bookID),
      createdAt: defaultDate,
      lastUpdated: defaultDate,
    };

    updateBook(book);

  };

  const openModal = () => setOpened(true);
  const closeModal = () => setOpened(false);

  console.log('RHF Errors: ', errors);
  return (
    <section className="bg-white dark:bg-gray-900 my-20">

      <div className="text-left px-4 py-8 md:pl-24 mx-auto max-w-2xl lg:py-16">
        <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">Edit Book</h2>
        <form className="grid gap-4 sm:grid-cols-2 sm:gap-6" onSubmit={handleSubmit(onSubmit)}>
          <div className="sm:col-span-2">
            <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white" htmlFor="title">Title<span className="text-red-600 ml-px">*</span></label>
            <input className="bg-gray-50 border border-gray-00 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" id="title" {...register('title')} />
          </div>
          <div className="block sm:col-span-2">
            <label className="block mb-2 text-sm font-medium text-gray-900 dark:text-white" htmlFor="subtitle">Subtitle</label>
            <input className="bg-gray-50 border border-gray-00 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" id="subtitle" {...register('subtitle')} />
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
                  <button type="button" onClick={() => authorFields.length > 1 && removeAuthor(index)}  className="flex flex-row justify-between items-center bg-dark-clay m-4">
                    <IoClose size={20}/>
                  </button>
                </div>
              ))}
              <button type="button" onClick={() => appendAuthor('')} className="flex flex-row justify-between items-center bg-dark-clay m-4">
                <IoAddOutline size={20} className="mr-1"/>
                Add Author
              </button>
            </div>
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
                    <button type="button" onClick={() => genreFields.length > 1 && removeGenre(index)} className="flex flex-row justify-between items-center bg-dark-clay m-4">
                      <IoClose size={20}/>
                    </button>
                  </div>
                ))}
              <button type="button" onClick={() => appendGenre('')} className="flex flex-row justify-between items-center bg-dark-clay m-4">
                <IoAddOutline size={20} className="mr-1"/>
                Add Genre
              </button>
            </div>
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
          </div>

          {/* Publish Date */}
          <div className="sm:col-span-2">
            <label htmlFor="publishDate" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Publish Date<span className="text-red-600 ml-px">*</span></label>
            <input id="publishDate" {...register('publishDate')} className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500"/>
          </div>

          <div className="w-full">
            <label htmlFor="isbn10" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">ISBN-10<span className="text-red-600 ml-px">*</span></label>
            <input id="isbn10" {...register('isbn10')} className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />
          </div>

          <div className="w-full">
            <label htmlFor="isbn13" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">ISBN-13<span className="text-red-600 ml-px">*</span></label>
            <input id="isbn13" {...register('isbn13')} className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />
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
          </div>

          {/* Language */}
          <div className="w-full">
            <label htmlFor="language" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Language<span className="text-red-600 ml-px">*</span></label>
            <input id="language" {...register('language')} className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />
          </div>

          {/* Page Count */}
          <div className="w-full">
            <label htmlFor="pageCount" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Page Count<span className="text-red-600 ml-px">*</span></label>
            <input id="pageCount" type="number" {...register('pageCount', { valueAsNumber: true })} className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />
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
            <label htmlFor="description" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Description<span className="text-red-600 ml-px">*</span></label>
            <textarea id="description" rows={4} {...register('description')} className="block p-2.5 w-full text-sm text-gray-900 bg-gray-50 rounded border border-gray-300 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />
          </div>

          {/* Notes */}
          <div className="sm:col-span-2">
            <label htmlFor="notes" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Notes<span className="text-red-600 ml-px">*</span></label>
            <textarea id="notes" rows={4} {...register('notes')} className="block p-2.5 w-full text-sm text-gray-900 bg-gray-50 rounded border border-gray-300 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />
          </div>

          <button type="submit">Update Book</button>
          <button type="button" onClick={openModal} className="border-red-500 text-red-500 hover:text-white hover:bg-red-600 focus:ring-red-900">Delete Book</button>
          <Modal opened={opened} onClose={closeModal} title="Danger zone">
              <div className="flex items-center justify-center">
                <IoIosWarning size={30} />
              </div>
              <h3 className="flex items-center justify-center text-lg">Are you sure that you want to delete this book?</h3>
              <p className="flex items-center justify-center mb-5">This action cannot be undone.</p>
              <button onClick={closeModal} className="flex flex-row justify-between items-center bg-transparent mr-1 w-full mb-3 lg:mb-0">
                <span>Cancel</span>
              </button>
              <button className="flex flex-row justify-between items-center bg-transparent mr-1 w-full text-white bg-red-600 hover:bg-red-800 focus:ring-red-800 mb-3 lg:mb-0">
                <span>Yes, I want to delete this book</span>
                <MdDeleteForever size={30}/>
              </button>
            </Modal>
        </form>
      </div>
    </section>
  );
};

export default EditBook;
