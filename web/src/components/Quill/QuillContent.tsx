import React from 'react';
import DOMPurify from 'dompurify';

interface DeltaOp {
  insert: string;
  attributes?: {
    bold?: boolean;
    italic?: boolean;
    underline?: boolean;
    header?: number;
    list?: 'ordered' | 'bullet';
    align?: 'center' | 'right' | 'justify';
    link?: string;
    blockquote?: boolean;
    code?: boolean;
    script?: 'sub' | 'super';
    color?: string;
    background?: string;
  };
}

interface QuillContentProps {
  content: {
    ops: DeltaOp[] | null;
  };
}

const QuillContent: React.FC<QuillContentProps> = ({ content }) => {
  const deltaToHtml = (ops: DeltaOp[] | null): string => {
    if (!ops) return '';

    let html = '';
    let listType: string | null = null;
    let inBlockquote = false;

    ops.forEach((op) => {
      let text = op.insert;
      const attrs = op.attributes || {};

      // Handle line breaks
      if (text === '\n') {
        if (!attrs.header && !attrs.list && !attrs.blockquote) {
          text = '<br>';
        }
      }

      // Apply inline formatting
      if (attrs) {
        if (attrs.bold) text = `<strong>${text}</strong>`;
        if (attrs.italic) text = `<em>${text}</em>`;
        if (attrs.underline) text = `<u>${text}</u>`;
        if (attrs.code) text = `<code>${text}</code>`;
        if (attrs.script === 'sub') text = `<sub>${text}</sub>`;
        if (attrs.script === 'super') text = `<sup>${text}</sup>`;
        if (attrs.link) text = `<a href="${attrs.link}" target="_blank" rel="noopener noreferrer">${text}</a>`;
        if (attrs.color) text = `<span style="color: ${attrs.color}">${text}</span>`;
        if (attrs.background) text = `<span style="background-color: ${attrs.background}">${text}</span>`;

        // Block-level formatting
        if (attrs.header) {
          text = `<h${attrs.header}>${text}</h${attrs.header}>`;
        } else if (attrs.blockquote && !inBlockquote) {
          text = `<blockquote>${text}`;
          inBlockquote = true;
        } else if (!attrs.blockquote && inBlockquote) {
          text = `</blockquote>${text}`;
          inBlockquote = false;
        } else if (attrs.list) {
          if (listType !== attrs.list) {
            if (listType) html += `</${listType}>`;
            listType = attrs.list === 'ordered' ? 'ol' : 'ul';
            html += `<${listType}>`;
          }
          text = `<li>${text}</li>`;
        } else if (listType) {
          html += `</${listType}>`;
          listType = null;
        }

        if (attrs.align) {
          text = `<div style="text-align: ${attrs.align}">${text}</div>`;
        }
      }

      html += text;
    });

    // Close any remaining open tags
    if (listType) html += `</${listType}>`;
    if (inBlockquote) html += '</blockquote>';

    return html;
  };

  const rawHtml = deltaToHtml(content.ops);
  const sanitizedHtml = DOMPurify.sanitize(rawHtml);

  return <div className="prose dark:prose-invert max-w-none" dangerouslySetInnerHTML={{ __html: sanitizedHtml }} />;
};

export default QuillContent;
