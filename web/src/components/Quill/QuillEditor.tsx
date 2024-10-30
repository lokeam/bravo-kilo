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

const QuillEditor: React.FC<QuillEditorProps> = ({
  value,
  onChange,
  placeholder,
  onError,
  readOnly = false
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const quillRef = useRef<Quill | null>(null);

  // Initialize Quill
  useEffect(() => {
    if (!containerRef.current) return;

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
            ['bold', 'italic', 'underline', 'strike'],
            ['blockquote', 'code-block'],
            [{ 'list': 'ordered'}, { 'list': 'bullet' }],
            [{ 'indent': '-1'}, { 'indent': '+1' }],
            ['clean']
          ]
        }
      });

      // Set initial value
      const initialDelta = value instanceof Delta ? value : new Delta(value.ops);
      quill.setContents(initialDelta, 'silent');

      // Set up change handler
      quill.on('text-change', (delta, oldDelta, source) => {
        if (source === 'user') {
          onChange(quill.getContents());
        }
      });

      // Set readonly state
      quill.enable(!readOnly);

      quillRef.current = quill;

      // Cleanup
      return () => {
        quill.off('text-change');
        container.innerHTML = '';
        quillRef.current = null;
      };
    } catch (error) {
      onError?.(error instanceof Error ? error : new Error('Error initializing Quill'));
    }
  }, []); // Empty dependency array - only run on mount/unmount

  // Handle value updates
  useEffect(() => {
    if (!quillRef.current) return;

    try {
      const quill = quillRef.current;
      const newDelta = value instanceof Delta ? value : new Delta(value.ops);
      const currentDelta = quill.getContents();

      if (JSON.stringify(currentDelta) !== JSON.stringify(newDelta)) {
        quill.setContents(newDelta, 'silent');
      }
    } catch (error) {
      onError?.(error instanceof Error ? error : new Error('Error updating Quill contents'));
    }
  }, [value]);

  return <div ref={containerRef} />;
};

export default QuillEditor;