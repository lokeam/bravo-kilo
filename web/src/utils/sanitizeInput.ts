import DOMPurify from 'dompurify';

/**
 * Sanitize all input fields that are strings or arrays of strings, with custom handling for specific fields
 * @param data - the form data object
 * @param fieldsToSanitize - array of field names that require sanitization
 * @param fieldsToTrim - array of field names that should be trimmed
 * @returns a new object with sanitized and trimmed fields
 */

export function sanitizeFormData<T>(
  data: T,
  fieldsToSanitize: (keyof T)[],
  fieldsToTrim: (keyof T)[]
): T {
  const sanitizedData = { ...data };

  // Sanitize fields
  for (const field of fieldsToSanitize) {
    const value = sanitizedData[field];

    if (typeof value === 'string') {
      sanitizedData[field] = DOMPurify.sanitize(value) as T[keyof T];
    }

    if (Array.isArray(value)) {
      sanitizedData[field] = value.map(item => {
        typeof item === 'string' ? DOMPurify.sanitize(item) : item
      }) as T[keyof T];
    }
  }

  // Trim any fields
  for (const field of fieldsToTrim) {
    const value = sanitizedData[field];

    if (typeof value === 'string') {
      sanitizedData[field] = value.trim() as T[keyof T];
    }
  }

  return sanitizedData;
}
