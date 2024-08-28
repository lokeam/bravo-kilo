import { useState } from 'react';
import { useAuth } from '../AuthContext';
import { useQuery } from '@tanstack/react-query';
import useDebounce from '../../hooks/useDebounceLD';
import { exportUserBooks } from '../../service/apiClient.service';

const SettingsItemExportBtn = () => {
  const { user } = useAuth();
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [ error, setError] = useState<string | null>(null);
  const userID = parseInt(user.id || 0, 10);

  // Assume useQuery is used to cache the user's books
  const booksQueryKey = ['userBooks', userID];
  const { data: cachedBooks } = useQuery({
    queryKey: booksQueryKey,
    queryFn: () => {}, // Fetching logic for cached books
    enabled: false,
  });

  // Check if the user has books cached
  const hasBooks = cachedBooks && cachedBooks.length > 0;


  // Handle the button click
  const handleExport = async () => {
    setIsLoading(true);
    setError(null);
    try {
      await exportUserBooks(userID);
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
        className="h-11 justify-stretch"
        disabled={!hasBooks || isLoading}
        onClick={debouncedExport}
      >
        {isLoading ? 'Exporting...' : hasBooks ? 'Export library data as CSV' : 'No data to export'}
      </button>
      {error && <p className="text-red-500 mt-2">{error}</p>}
    </div>
  );
};

export default SettingsItemExportBtn;
