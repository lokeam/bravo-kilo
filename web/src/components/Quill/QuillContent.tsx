import React from 'react';
import DOMPurify from 'dompurify';

interface DeltaOp {
  insert: string;
  attributes?: Record<string, any>;
}

interface QuillContentProps {
  content: {
    ops: DeltaOp[] | null;
  };
}

const QuillContent: React.FC<QuillContentProps> = ({ content }) => {
  const deltaToHtml = (ops: DeltaOp[] | null): string => {
    if (!ops) return ''; // Return empty string if ops is null

    let html = '';
    let listType: string | null = null;

    ops.forEach((op) => {
      let text = op.insert;

      if (op.attributes) {
        if (op.attributes.bold) text = `<strong>${text}</strong>`;
        if (op.attributes.italic) text = `<em>${text}</em>`;
        if (op.attributes.underline) text = `<u>${text}</u>`;

        if (op.attributes.header) {
          text = `<h${op.attributes.header}>${text}</h${op.attributes.header}>`;
        } else if (op.attributes.list) {
          if (listType !== op.attributes.list) {
            if (listType) html += `</${listType}>`;
            listType = op.attributes.list === 'ordered' ? 'ol' : 'ul';
            html += `<${listType}>`;
          }
          text = `<li>${text}</li>`;
        } else if (listType) {
          html += `</${listType}>`;
          listType = null;
        }
      } else if (text === '\n') {
        text = '<br>';
      }

      html += text;
    });

    if (listType) html += `</${listType}>`;

    return html;
  };

  const rawHtml = deltaToHtml(content.ops);
  const sanitizedHtml = DOMPurify.sanitize(rawHtml);

  return <div dangerouslySetInnerHTML={{ __html: sanitizedHtml }} />;
};

export default QuillContent;