import { useState, useMemo } from 'react';

interface FormatPublishDateResult {
  formattedDate: string;
  dateWarning: string | null;
}

export const useFormatPublishDate = (publishDate: string | undefined): FormatPublishDateResult => {
  const [dateWarning, setWarning] = useState<string | null>(null);

  const { formattedDate, warning } = useMemo(() => {
    let formattedDate = '';
    let warning: string | null = null;

    if (!publishDate) {
      const currentYear = new Date().getFullYear();
      warning = "Google Books didn't give us a valid publish date. Please double check the publish date.";
      formattedDate = `${currentYear}-01-01`;
    } else {
      const dateRegex = /^\d{4}(-\d{2}(-\d{2})?)?$/;
      if (!dateRegex.test(publishDate)) {
        warning = "Google Books didn't give us a valid publish date. Please double check the publish date.";
        formattedDate = `${new Date().getFullYear()}-01-01`;
      } else {
        const parts = publishDate.split('-');
        if (parts.length === 1) {
          warning = "Google Books only gave us the publish year. Please double check this book's publish date.";
          formattedDate = `${parts[0]}-01-01`;
        } else if (parts.length === 2) {
          warning = "Google Books only gave us the publish year and month. Please double check this book's publish date.";
          formattedDate = `${parts[0]}-${parts[1]}-01`;
        } else {
          formattedDate = publishDate;
        }
      }
    }

    return { formattedDate, warning };
  }, [publishDate]);

  if (warning !== dateWarning) {
    setWarning(warning);
  }

  return { formattedDate, dateWarning };
};
