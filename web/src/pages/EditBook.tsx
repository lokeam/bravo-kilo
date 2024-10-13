import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { SubmitHandler } from 'react-hook-form';
import Markdown from 'react-markdown';
import DOMPurify from 'dompurify';
import PageWithErrorBoundary from '../components/ErrorMessages/PageWithErrorBoundary';

import Loading from '../components/Loading/Loading';
import BookForm from '../components/BookForm/BookForm';
import Modal from '../components/Modal/ModalRoot';
import BookSummaryBtn from '../components/BookSummaryBtn/BookSummaryBtn';
import { BookFormData } from '../types/api';

import useUpdateBook from '../hooks/useUpdateBook';
import useDeleteBook from '../hooks/useDeleteBook';
import useFetchBookById from '../hooks/useFetchBookById';
import useStore from '../store/useStore';

import { IoIosWarning } from "react-icons/io";
import { MdDeleteForever } from "react-icons/md";
import { RiFileCopyLine } from "react-icons/ri";

const EditBook = () => {
  // Delete Modal state
  const [opened, setOpened] = useState(false);

  // AI Preview + Modal state
  const [aiSummaryPreview, setAiSummaryPreview] = useState("");
  const [isPreviewModalOpen, setIsPreviewModalOpen] = useState(false);
  const [isManualTrigger, setIsManualTrigger] = useState(false);

  const navigate = useNavigate();
  const { bookID } = useParams();
  const { showSnackbar } = useStore();

  const { data: book, isLoading: isFetchLoading, isError } = useFetchBookById(bookID as string, !!bookID);
  const { updateBook, isLoading: isUpdateLoading, LoadingComponent: UpdateLoadingComponent } = useUpdateBook(bookID as string);
  const { deleteBook, isLoading: isDeleteLoading, LoadingComponent: DeleteLoadingComponent } = useDeleteBook();

  if (isFetchLoading) return <Loading />;
  if (isError || !book) return <div>Error loading book data</div>;

  // Form Submittal
  const handleUpdateBook: SubmitHandler<BookFormData> = async (data) => {
    console.log(`Form submitted with data ${data}`);
    const defaultDate = new Date().toISOString();

    const updatedBook = {
      ...data,
      id: Number(bookID),
      createdAt: defaultDate,
      lastUpdated: defaultDate,
      isbn10: data.isbn10 || '',
      isbn13: data.isbn13 || '',
      authors: data.authors.map((authorObj) => authorObj.author.trim()).filter((author) => author !== ''),
      genres: data.genres.map((genreObj) => genreObj.genre.trim()).filter((genre) => genre !== ''),
      tags: data.tags.map((tagObj) => tagObj.tag.trim()).filter((tag) => tag !== ''),
    };

    try {
      await updateBook(updatedBook);
      showSnackbar('Book updated successfully', 'updated');
      navigate('/library');
    } catch (error) {
      console.error('Error updating book:', error);
      showSnackbar('Failed to update book. Please try again.', 'error');
    }
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
  const handleDelete = async (event: React.MouseEvent<HTMLButtonElement>) => {
    event?.preventDefault();
    try {
      await deleteBook(bookID as string);
      showSnackbar('Book deleted successfully', 'removed');
      navigate('/library');
    } catch (error) {
      console.error('Error deleting book:', error);
      showSnackbar('Failed to delete book. Please try again.', 'error');
    } finally {
      closeModal();
    }
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

  const renderAISummaryBtn = (
    <div className="grid w-full gap-6 lgMobile:grid-cols-3 pt-2">
      <BookSummaryBtn
        title={book?.title || ''}
        authors={book?.authors || ['']}
        setAiSummaryPreview={setAiSummaryPreview}
        setIsManualTrigger={setIsManualTrigger}
        openPreviewModal={openPreviewModal}
        isPreviewModalOpen={isPreviewModalOpen}
        isManualTrigger={isManualTrigger}
      />
    </div>
  );

  return (
    <PageWithErrorBoundary fallbackMessage="Error loading edit page">
      <section className="editBook bg-white dark:bg-black min-h-screen bg-cover relative flex flex-col items-center place-content-around px-5 antialiased mdTablet:pr-5 mdTablet:ml-24">
        <div className="text-left text-dark max-w-screen-mdTablet pb-24 md:pb-4 flex flex-col relative w-full">
          <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">Edit Book</h2>
          <BookForm
            onSubmit={handleUpdateBook}
            initialData={book}
            isEditMode={true}
            onDelete={openModal}
            renderAISummaryBtn={renderAISummaryBtn}
            isLoading={isUpdateLoading}
          />

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
