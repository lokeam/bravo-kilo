import { useEffect, useRef } from 'react';
import { debounce } from 'lodash';

export default function useDebounce<T extends (...args: any[]) => void>(callback: T, delayInMs: number) {
  const callbackRef = useRef(callback);

  // Update current callback ref upon change
  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  // Persist debounced fn in ref
  const debouncedCallback = useRef(
    debounce((...args: Parameters<T>) => {
      callbackRef.current(...args);
    }, delayInMs)
  ).current;

  // Clean up debounced fn on unmount
  useEffect(() => {
    return () => debouncedCallback.cancel();
  }, [debouncedCallback]);

  return debouncedCallback;
}
