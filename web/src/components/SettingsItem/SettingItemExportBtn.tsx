import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useUser } from '../../hooks/useUser';
import useDebounce from '../../hooks/useDebounceLD';
import { exportUserBooks } from '../../service/apiClient.service';

function SettingsItemExportBtn() {
  const { data: user} = useUser();
  const userID = user?.id;
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [ error, setError] = useState<string | null>(null);

  // Assume useQuery is used to cache the user's books
  const booksQueryKey = ['books', userID];
  const { data: cachedBooks } = useQuery({
    queryKey: booksQueryKey,
    queryFn: async () => {
      // Fetch user books here
      return [];
    },
    enabled: false,
  });

  // Check if the user has books cached
  const hasBooks = cachedBooks && cachedBooks.length > 0;

  console.log('************');
  console.log('userID: ', userID);
  console.log('checking hasBooks: ', cachedBooks);
  console.log('checking hasBooks: ', hasBooks);

  // Handle the button click
  const handleExport = async () => {
    setIsLoading(true);
    setError(null);
    try {
      if (userID !== undefined) {
        await exportUserBooks(userID);
      } else {
        throw new Error("User ID is not available");
      }
    } catch(error) {
      setError("Failed to export books. Please try again later.");
    } finally {
      setIsLoading(false);
    }
  };

  const debouncedExport = useDebounce(handleExport, 2000);

  return (
    <div className="grid w-full">
      <button
        className="bg-gray-200 dark:bg-gray-800 dark:text-white border-2 dark:border-2 transition duration-500 ease-in-out hover:border-vivid-blue dark:hover:border-vivid-blue h-11 justify-stretch"
        disabled={!hasBooks || isLoading}
        onClick={debouncedExport}
      >
        {isLoading ? 'Exporting...' : hasBooks ? 'Export library data as CSV' : 'No data to export'}
      </button>
      {error && <p className="text-red-500 mt-2">{error}</p>}
    </div>
  );
}

export default SettingsItemExportBtn;
