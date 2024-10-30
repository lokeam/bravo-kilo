import React from 'react';
import DOMPurify from 'dompurify';
import { deltaToHtml } from '../../utils/formatRichText';
import { DeltaOp } from '../../types/quill';

interface QuillContentProps {
  content: {
    ops: DeltaOp[] | null;
  };
}

const QuillContent: React.FC<QuillContentProps> = ({ content }) => {
  const rawHtml = deltaToHtml(content.ops);
  const sanitizedHtml = DOMPurify.sanitize(rawHtml);

  return <div className="prose dark:prose-invert max-w-none" dangerouslySetInnerHTML={{ __html: sanitizedHtml }} />;
};

export default QuillContent;
