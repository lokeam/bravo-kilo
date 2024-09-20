import { useEffect, useState } from 'react';
import useDebounce from '../../hooks/useDebounceLD';
import useGeminiPrompt from '../../hooks/useGeminiPrompt';
import Loading from '../Loading/Loading';

interface BookSummaryBtnProps {
  title: string;
  authors: string[];
  setAiSummaryPreview: (summary: string) => void;
  setIsManualTrigger: (isManualTrigger: boolean) => void;
  openPreviewModal: () => void;
  isPreviewModalOpen: boolean;
  isManualTrigger: boolean;
}

const BookSummaryBtn = (
  {
    title,
    authors,
    setAiSummaryPreview,
    setIsManualTrigger,
    openPreviewModal,
    isPreviewModalOpen,
    isManualTrigger
  }: BookSummaryBtnProps) => {
  const [error, setError] = useState<string | null>(null);
  const prompt = `
  Ignore all prior instructions.

  Act as an expert on summarization, outlining and structuring. Your style of writing should be informative and logical.

  Provide me with a detailed summary of the book titled "${title}" by ${authors.join(", ")}.
  This summary should use clear and concise language in order to make it easy to understand.
  Generate a detailed summary of the book titled "${title}" by ${authors.join(", ")}.

  Do not introduce yourself, do not remind me what I asked you for. Do not apologize. Do not self-reference.

  Generate the output in markdown format. Make sure your response is no longer than four paragraphs `;

  const { data: promptResponse, isLoading, isError, refetch } = useGeminiPrompt(prompt);

  useEffect(() => {
    if (
      promptResponse &&
      !isLoading &&
      !isPreviewModalOpen &&
      isManualTrigger
    ) {
      const formattedResponse = promptResponse.parts[0].replace(/['‘’"“”]/g, '');
      //console.log('checking formatted response: ', formattedResponse);

      setAiSummaryPreview(formattedResponse);
      openPreviewModal();
    }
    if (isError) {
      setError("Failed to generate book summary. Please try again later.")
    }
  }, [promptResponse, isLoading, isPreviewModalOpen, setAiSummaryPreview, openPreviewModal, isError, isManualTrigger]);

  const handleClick = useDebounce((event: React.MouseEvent<HTMLButtonElement>) => {
    console.log('Gemini button clicked');
    setIsManualTrigger(true);
    event.preventDefault();
    setError(null);
    refetch();
  }, 2000);

  //console.log('Book Summary Button - prompt to send: ', prompt);

  return (
    <>
      <button
        className="book_summary p-[1.75px] transtion group flex h-[45.19px] w-auto items-center justify-center rounded-lg bg-gradient-to-r from-seljuk from-11.63% via-lilac via-40.43% to-carmine to-68.07% text-white duration-300 hover:bg-gray-700"
        onClick={handleClick}
        type="button">
        <span className="book_summary_bg flex h-full w-full items-center justify-center rounded-lg bg-black transition duration-300 ease-in-out hover:bg-gray-700">
          {isLoading ? 'Generating...' : 'Summarize Book with AI'}
        </span>

      </button>
      { isLoading && <Loading /> }
      {
        error && (
        <div className="text-red-500 mt-2">
          {error}
          <button
            className="ml-2 text-blue-500"
            onClick={handleClick}
          >
            Retry
          </button>
        </div>)
      }
    </>
  )
};

export default BookSummaryBtn;
