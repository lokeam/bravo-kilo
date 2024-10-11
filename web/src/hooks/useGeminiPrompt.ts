import { useQuery } from '@tanstack/react-query';
import { geminiQueryAPI } from '../service/apiClient.service';

const useGeminiPrompt = (prompt: string) => {

  return useQuery({
    queryKey: ['prompt', prompt],
    queryFn: async () => {
      const promptResponse = await geminiQueryAPI(prompt);
      console.log('useGeminiPrompt.ts: response:', promptResponse);
      return promptResponse;
    },
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 5,
    enabled: false,
  });
};

export default useGeminiPrompt;
