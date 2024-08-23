import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Controller, useForm, useFieldArray, SubmitHandler } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

import useUpdateBook from '../hooks/useUpdateBook';
import useDeleteBook from '../hooks/useDeleteBook';
import useFetchBookById from '../hooks/useFetchBookById';

import Modal from '../components/Modal/ModalRoot';
import { Book } from './Library';

import { IoClose } from 'react-icons/io5';
import { IoAddOutline } from 'react-icons/io5';
import { IoIosWarning } from "react-icons/io";
import { MdDeleteForever } from "react-icons/md";
import { deleteBook } from '../service/apiClient.service';


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
  imageLink: z.string().min(1, 'Please enter an image link'),
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
  const { mutate: deleteBook } = useDeleteBook();
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
      imageLink: data.imageLink.trim(),
      createdAt: defaultDate,
      lastUpdated: defaultDate,
    };

    updateBook(book);
    navigate('/library');
  };

  const openModal = () => setOpened(true);
  const closeModal = () => setOpened(false);
  const handleDelete = (event: React.MouseEvent<HTMLButtonElement>) => {
    event?.preventDefault;
    deleteBook(bookID as string)
  };

  console.log('RHF Errors: ', errors);

  return (
    <section className="bg-black relative flex flex-col items-center place-content-around px-5 antialiased mdTablet:pr-5 mdTablet:ml-24 h-screen">
      <div className="text-left max-w-screen-mdTablet py-24 md:pb-4 flex flex-col relative w-full">
        <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">Edit Book</h2>
        <form className="grid gap-4 grid-cols-2 sm:gap-6" onSubmit={handleSubmit(onSubmit)}>

          {/* Title */}
          <div className="block col-span-2 mdTablet:col-span-1">
            <label className="block mb-2 text-base font-medium text-gray-900 dark:text-white" htmlFor="title">Title<span className="text-red-600 ml-px">*</span></label>
            <input className="bg-maastricht border border-gray-00 text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" id="title" {...register('title')} />
            {errors.title && <p className="text-red-500">{errors.title.message}</p>}
          </div>

          {/* Subtitle */}
          <div className="block col-span-2 mdTablet:col-span-1">
            <label className="block mb-2 text-base font-medium text-gray-900 dark:text-white" htmlFor="subtitle">Subtitle (if applicable)</label>
            <input className="bg-maastricht border border-gray-00 text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" id="subtitle" {...register('subtitle')} />
            {errors.subtitle && <p className="text-red-500">{errors.subtitle.message}</p>}
          </div>

          {/* Authors Field Array */}
          <div className="block col-span-2">
            <label className="block mb-2 text-base font-medium text-gray-900 dark:text-white">Authors<span className="text-red-600 ml-px">*</span></label>
            <div className="border border-cadet-gray rounded p-4">
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
            <div className="border border-cadet-gray rounded p-4">
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
            <div className="border border-cadet-gray rounded p-4">
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
            <input id="publishDate" {...register('publishDate')} className="bg-maastricht border border-gray-300 text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5  dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500"/>
            {errors.publishDate && <p className="text-red-500">{errors.publishDate.message}</p>}
          </div>

          {/* ISBN10 */}
          <div className="block col-span-2">
            <label htmlFor="isbn10" className="block mb-2  text-base  font-medium text-gray-900 dark:text-white">ISBN-10<span className="text-red-600 ml-px">*</span></label>
            <input id="isbn10" {...register('isbn10')} className="bg-maastricht border border-gray-300 text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5  dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />
            {errors.isbn10 && <p className="text-red-500">{errors.isbn10.message}</p>}
          </div>

          {/* ISBN13 */}
          <div className="block col-span-2">
            <label htmlFor="isbn13" className="block mb-2  text-base  font-medium text-gray-900 dark:text-white">ISBN-13<span className="text-red-600 ml-px">*</span></label>
            <input id="isbn13" {...register('isbn13')} className="bg-maastricht border border-gray-300 text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5  dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />
            {errors.isbn13 && <p className="text-red-500">{errors.isbn13.message}</p>}
          </div>

          {/* Formats */}
          <div className="col-span-2">
            <label className="block mb-2 text-base  font-medium text-gray-900 dark:text-white">Formats (select all that apply)<span className="text-red-600 ml-px">*</span></label>
              <ul className="grid w-full gap-6 lgMobile:grid-cols-3">
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
            <label htmlFor="language" className="block mb-2 text-base  font-medium text-gray-900 dark:text-white">Language<span className="text-red-600 ml-px">*</span></label>
            <input id="language" {...register('language')} className="bg-maastricht border border-gray-300 text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5  dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />
            {errors.language && <p className="text-red-500">{errors.language.message}</p>}
          </div>

          {/* Page Count */}
          <div className="col-span-2">
            <label htmlFor="pageCount" className="block mb-2 text-base font-medium text-gray-900 dark:text-white">Page Count<span className="text-red-600 ml-px">*</span></label>
            <input id="pageCount" type="number" {...register('pageCount', { valueAsNumber: true })} className="bg-maastricht border border-gray-300 text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5  dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />
            {errors.pageCount && <p className="text-red-500">{errors.pageCount.message}</p>}
          </div>

          {/* Image Links Field Array */}
          <div className="col-span-2">
            <label className="block mb-2 text-base font-medium text-gray-900 dark:text-white">Image Link<span className="text-red-600 ml-px">*</span></label>
            <input className="bg-maastricht border border-gray-00 text-gray-900 text-base rounded focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" id="imageLink" {...register('imageLink')} />
            {errors.imageLink && <p className="text-red-500">{errors.imageLink.message}</p>}
          </div>

          {/* Description */}
          <div className="col-span-2">
            <label htmlFor="description" className="block mb-2 text-base font-medium text-gray-900 dark:text-white">Description<span className="text-red-600 ml-px">*</span></label>
            <textarea id="description" rows={4} {...register('description')} className="block p-2.5 w-full text-base text-gray-900 bg-maastricht rounded border border-gray-300 focus:ring-primary-500 focus:border-primary-500  dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />
            {errors.description && <p className="text-red-500">{errors.description.message}</p>}
          </div>

          {/* Notes */}
          <div className="col-span-2">
            <label htmlFor="notes" className="block mb-2 text-base font-medium text-gray-900 dark:text-white">Notes<span className="text-red-600 ml-px">*</span></label>
            <textarea id="notes" rows={4} {...register('notes')} className="block p-2.5 w-full text-base text-gray-900 bg-maastricht rounded border border-gray-300 focus:ring-primary-500 focus:border-primary-500  dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" />
            {errors.notes && <p className="text-red-500">{errors.notes.message}</p>}
          </div>

          <button className="bg-majorelle hover:bg-hepatica" type="submit">Update Book</button>
          <button type="button" onClick={openModal} className="border-red-500 text-red-500 hover:text-white hover:bg-red-600 focus:ring-red-900">Delete Book</button>
        </form>
        <Modal opened={opened} onClose={closeModal} title="Danger zone">
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
      </div>
    </section>
  );
};

export default EditBook;
