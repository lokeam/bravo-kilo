import React, { useEffect, useRef } from 'react';
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
  readOnly?: boolean;
}

// Typeguard to check if obj has ops property
function hasOps(obj: any): obj is { ops: any[] } {
  return obj && typeof obj === 'object' && 'ops' in obj;
}


const QuillEditor: React.FC<QuillEditorProps> = ({
  value,
  onChange,
  placeholder,
  onError,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const quillRef = useRef<Quill | null>(null);
  const initializedRef = useRef(false);

  const parseValue = (val: any): Delta => {
    console.log('Parsing value:', val);

    try {
      if (!val) {
        return new Delta();
      }

      if (val instanceof Delta) {
        console.log('Value is Delta instance:', val);

        // Handle nested Delta structure
        const insert = val.ops?.[0].insert;
        if (
          insert &&
          typeof insert === 'object' &&
          hasOps(insert)
        ) {
          console.log('Found nested Delta structure');
          const nestedOps = insert.ops;

          // Handle stringified content in nested ops
          if (typeof nestedOps[0]?.insert === 'string') {
            try {
              const parsed = JSON.parse(nestedOps[0].insert);
              if (parsed.ops === null) {
                return new Delta([{ insert: '\n' }]);
              }
              return new Delta(parsed.ops || []);
            } catch {
              return new Delta([{ insert: nestedOps[0].insert }]);
            }
          }

          return new Delta(nestedOps);
        }

        // If ops is null or empty, return empty Delta
        if (!val.ops || val.ops.length === 0) {
          return new Delta([{ insert: '\n' }]);
        }

        return val;
      }

      // Handle stringified Delta
      if (typeof val === 'string') {
        try {
          const parsed = JSON.parse(val);
          console.log('Parsed string value:', parsed);
          if (!parsed.ops || parsed.ops === null) {
            return new Delta([{ insert: '\n' }]);
          }
          return new Delta(parsed.ops || [{ insert: val }]);
        } catch {
          return new Delta([{ insert: val }]);
        }
      }

      // Handle Delta-like object
      if (typeof val === 'object' && 'ops' in val) {
        console.log('Handling Delta-like object:', val);

        // Handle nested Delta structure
        if (val.ops?.[0]?.insert?.ops) {
          const nestedOps = val.ops[0].insert.ops;
          return new Delta(nestedOps);
        }

        // If ops is null or empty string, return empty Delta
        if (!val.ops || val.ops === null || val.ops === '') {
          return new Delta([{ insert: '\n' }]);
        }

        // Handle case where ops might be stringified
        if (typeof val.ops === 'string') {
          try {
            const parsed = JSON.parse(val.ops);
            console.log('Parsed ops string:', parsed);
            if (!parsed || parsed === null) {
              return new Delta([{ insert: '\n' }]);
            }
            return new Delta(parsed || []);
          } catch {
            return new Delta([{ insert: val.ops }]);
          }
        }

        // Handle normal ops array
        if (Array.isArray(val.ops)) {
          console.log('Handling ops array:', val.ops);
          return new Delta(val.ops);
        }

        return new Delta([{ insert: '\n' }]);
      }

      return new Delta([{ insert: '\n' }]);
    } catch (error) {
      console.error('Error parsing value:', error, 'Original value:', val);
      return new Delta([{ insert: '\n' }]);
    }
  };

  // Initialize Quill
  useEffect(() => {
    if (!containerRef.current || initializedRef.current) return;

    try {
      const container = containerRef.current;
      const editorContainer = container.appendChild(
        container.ownerDocument.createElement('div')
      );

      const quill = new Quill(editorContainer, {
        theme: 'snow',
        placeholder,
        modules: {
          toolbar: [
            [{ header: [1, 2, 3,false] }],
            ['bold', 'italic', 'underline', 'strike'],
            [{ color: [] }, { background: [] }],
            ['blockquote'],
            [{ 'list': 'ordered'}, { 'list': 'bullet' }],
            [{ 'indent': '-1'}, { 'indent': '+1' }],
            ['clean']
          ]
        }
      });

      // Set initial value
      const initialDelta = parseValue(value);
      quill.setContents(initialDelta, 'silent');

      // Set up change handler
      // @ts-expect-error
      quill.on('text-change', (delta, oldDelta, source) => {
        if (source === 'user') {
          const contents = quill.getContents();
          onChange(contents);
        }
      });

      quillRef.current = quill;
      initializedRef.current = true;

      return () => {
        quill.off('text-change');
        container.innerHTML = '';
        quillRef.current = null;
        initializedRef.current = false;
      };
    } catch (error) {
      onError?.(error instanceof Error ? error : new Error('Error initializing Quill'));
    }
  }, []);

  // Handle value updates
  useEffect(() => {
    if (!quillRef.current || !initializedRef.current) return;

    try {
      const quill = quillRef.current;
      const newDelta = parseValue(value);
      const currentDelta = quill.getContents();

      if (JSON.stringify(currentDelta) !== JSON.stringify(newDelta)) {
        quill.setContents(newDelta, 'silent');
      }
    } catch (error) {
      console.error('Error updating Quill contents:', error);
      onError?.(error instanceof Error ? error : new Error('Error updating Quill contents'));
    }
  }, [value]);

  return <div ref={containerRef} />;
};

export default QuillEditor;
