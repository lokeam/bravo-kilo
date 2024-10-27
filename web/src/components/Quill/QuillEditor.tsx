import React, { useEffect, useRef, useCallback } from 'react';
import Quill from 'quill';
import Delta from 'quill-delta';
import 'quill/dist/quill.snow.css';

interface QuillEditorProps {
  value: Delta | { ops: any[]};
  onChange: (content: Delta) => void;
  placeholder?: string;
}

const isEqualDelta = (a: Delta | { ops: any[] }, b: Delta | { ops: any[] }): boolean => {
  return JSON.stringify(a) === JSON.stringify(b);
};

const QuillEditor: React.FC<QuillEditorProps> = ({ value, onChange, placeholder }) => {
  const editorRef = useRef<HTMLDivElement>(null);
  const quillRef = useRef<Quill | null>(null);

  const initQuill = useCallback(() => {
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
        onChange(quill.getContents());
      });

      quillRef.current = quill;
    }
  }, [onChange, placeholder]);

  useEffect(() => {
    initQuill();

    return () => {
      if (quillRef.current) {
        quillRef.current.off('text-change');
      }
    };
  }, [initQuill]);

  useEffect(() => {
    if (quillRef.current) {
      const quill = quillRef.current;
      quill.off('text-change');
      quill.on('text-change', () => {
        const contents = quill.getContents();
        onChange(contents);
      });

      const newValue = value instanceof Delta ? value : new Delta(value.ops);
      if (!isEqualDelta(quill.getContents(), newValue)) {
        quill.setContents(newValue);
      }
    }
  }, [onChange, value]);

  return <div ref={editorRef} />;
};

export default QuillEditor;