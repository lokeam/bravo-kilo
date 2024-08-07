import { useState, useEffect } from "react";

export const useDebounce = (value: string, delayInMs: number) => {
  const [ debouncedValue, setDebouncedValue ] = useState<string>(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delayInMs);

    return () => {
      clearTimeout(handler)
    };
  }, [value, delayInMs]);

  return debouncedValue;
};
