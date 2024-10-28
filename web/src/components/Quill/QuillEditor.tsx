import React, { useEffect, useRef, useCallback, useState } from 'react';
import Quill from 'quill';
import Delta from 'quill-delta';
import 'quill/dist/quill.snow.css';

interface QuillDeltaOps {
  insert?: string | Record<string, unknown>;
  retain?: number;
  delete?: number;
  attributes?: Record<string, unknown>;
}

interface QuillEditorProps {
  value: Delta | { ops: QuillDeltaOps[] };
  onChange: (content: Delta) => void;
  placeholder?: string;
  onError?: (error: Error) => void;
}

// Utility function to validate Delta structure
const isValidDelta = (value: any): value is Delta | { ops: QuillDeltaOps[] } => {
  try {
    if (!value || typeof value !== 'object') return false;
    if (value instanceof Delta) return true;
    if (!Array.isArray(value.ops)) return false;

    return value.ops.every((op: any) => {
      if (typeof op !== 'object') return false;
      if (
        op.insert === undefined &&
        op.delete === undefined &&
        op.retain === undefined
      ) return false;

      // Additional type checking for insert property
      if (
        op.insert !== undefined &&
        typeof op.insert !== 'string' &&
        (typeof op.insert !== 'object' || op.insert === null)
      ) return false;


      return true;
    });
  } catch (error) {
    console.error('Delta validation error: ', error);
    return false;
  }
};

// Equality check for Delta objects
const isEqualDelta = (
  a: Delta | { ops: QuillDeltaOps[] },
  b: Delta | { ops: QuillDeltaOps[] }
): boolean => {
  try {
    return JSON.stringify(a) === JSON.stringify(b);
  } catch (error) {
    console.error('Delta comparison error: ', error);
    return false;
  }
};

// Safe Delta conversion utility
const safeDeltaConversion = (value: Delta | { ops: QuillDeltaOps[] }): Delta => {
  try {
    if (value instanceof Delta) return value;
    if (!isValidDelta(value)) {
      throw new Error('Invalid Delta structure');
    }
    return new Delta(value.ops as Delta['ops']);
  } catch (error) {
    console.error('Delta conversion error:', error);
    return new Delta(); // Return empty Delta as fallback
  }
};

const QuillEditor: React.FC<QuillEditorProps> = ({
  value,
  onChange,
  placeholder,
  onError,
}) => {
  const editorRef = useRef<HTMLDivElement>(null);
  const quillRef = useRef<Quill | null>(null);
  const [hasError, setHasError] = useState<boolean>(false);

  const handleError = useCallback((error: Error) => {
    setHasError(true);
    console.error('QuillEditor error: ', error);
    onError?.(error);
  }, [onError])


  const initQuill = useCallback(() => {
    try {
      if (editorRef.current && !quillRef.current) {
        const quill = new Quill(editorRef.current, {
          theme: 'snow',
          modules: {
            toolbar: [
              [{ 'header': [1, 2, 3, 4, false] }],
              ['bold', 'italic', 'underline'],
              [{ 'list': 'ordered'}, { 'list': 'bullet' }],
            ]
          },
          placeholder: placeholder,
        });

        quill.on('text-change', () => {
          try {
            const contents = quill.getContents();
            onChange(contents);
          } catch (error) {
            handleError(error instanceof Error ? error : new Error('Error handling text change'));
          }
        });

        quillRef.current = quill;
      }
    } catch (error) {
      handleError(error instanceof Error ? error : new Error('Error initializing Quill'));
    }
  }, [onChange, placeholder, handleError]);

  useEffect(() => {
    initQuill();

    return () => {
      if (quillRef.current) {
        quillRef.current.off('text-change');
      }
    };
  }, [initQuill]);


  useEffect(() => {
    try {
      initQuill();

      return () => {
        if (quillRef.current) {
          try {
            quillRef.current.off('text-change');
          } catch (error) {
            console.error('Error cleaning up Quill instance:', error);
          }
        }
      };
    } catch (error) {
      handleError(error instanceof Error ? error : new Error('Error in Quill initialization effect'));
    }
  }, [initQuill, handleError]);

  useEffect(() => {
    try {
      if (!isValidDelta(value)) {
        throw new Error('Invalid Delta value provided');
      }

      if (quillRef.current) {
        const quill = quillRef.current;

        // Remove old handler before adding new one to prevent duplicates
        quill.off('text-change');

        quill.on('text-change', () => {
          try {
            const contents = quill.getContents();
            onChange(contents);
          } catch (error) {
            handleError(error instanceof Error ? error : new Error('Error handling text change'));
          }
        });

        const newValue = safeDeltaConversion(value);
        if (!isEqualDelta(quill.getContents(), newValue)) {
          quill.setContents(newValue);
        }
      }
    } catch (error) {
      handleError(error instanceof Error ? error : new Error('Error updating Quill contents'));
    }
  }, [onChange, value, handleError]);


  if (hasError) {
    return (
      <div className="quill-error-state">
        <p>Error loading editor. Please try refreshing the page.</p>
        <div className="hidden" ref={editorRef} />
      </div>
    );
  }

  return <div ref={editorRef} />;
};

export default QuillEditor;
