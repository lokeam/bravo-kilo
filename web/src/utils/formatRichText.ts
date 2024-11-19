import { DeltaOp, Block } from '../types/quill';

export const deltaToHtml = (ops: DeltaOp[] | null): string => {
  if (!ops) return '';

  let html = '';
  let currentBlock: Block | null = null;
  let listType: string | null = null;
  let currentListType: 'bullet' | 'ordered' | null = null;
  let inBlockquote = false;
  let paragraphContent = '';

  const wrapBlock = (force = false) => {
    const content = currentBlock?.content.trim() || paragraphContent.trim();

    if (!content) {
      currentBlock = null;
      paragraphContent = '';
      return;
    }

    if (currentBlock) {
      switch (currentBlock.type) {
        case 'header':
          html += `<h${currentBlock.level} class="mb-4">${content}</h${currentBlock.level}>`;
          break;
        case 'blockquote':
          html += `<blockquote><p>${content}</p></blockquote>`;
          break;
        case 'list':
          if (!listType) {
            listType = currentListType === 'ordered' ? 'ol' : 'ul';
            html += `<${listType}>`;
          }
          html += `<li class="mb-1">${content}</li>`;
          break;
        default:
          html += `<p class="mb-2">${content}</p>`;
      }
      currentBlock = null;
    } else if (force) {
      html += `<p class="mb-2">${content}</p>`;
    }

    paragraphContent = '';
  };

  const processInlineFormatting = (text: string, attrs: DeltaOp['attributes'] = {}): string => {
    if (!text) return '';
    if (attrs.bold) text = `<strong>${text}</strong>`;
    if (attrs.italic) text = `<em>${text}</em>`;
    if (attrs.underline) text = `<u>${text}</u>`;
    if (attrs.code) text = `<code>${text}</code>`;
    if (attrs.script === 'sub') text = `<sub>${text}</sub>`;
    if (attrs.script === 'super') text = `<sup>${text}</sup>`;
    if (attrs.link) text = `<a href="${attrs.link}" target="_blank" rel="noopener noreferrer">${text}</a>`;
    if (attrs.color) text = `<span style="color: ${attrs.color}">${text}</span>`;
    if (attrs.background) text = `<span style="background-color: ${attrs.background}">${text}</span>`;
    return text;
  };

  ops.forEach((op, index) => {
    const attrs = op.attributes || {};
    const text = typeof op.insert === 'string' ? op.insert : '';
    const nextOp = index < ops.length - 1 ? ops[index + 1] : null;

    if (text === '\n') {
      if (attrs.header) {
        // Close any open list before header
        if (listType) {
          html += `</${listType}>`;
          listType = null;
          currentListType = null;
        }
        html += `<h${attrs.header} class="mb-4 text-2xl">${paragraphContent.trim()}</h${attrs.header}>`;
        paragraphContent = '';
      } else if (attrs.list) {
        const trimmedContent = paragraphContent.trim();
        if (trimmedContent) {
          // Only start a new list if the type changes
          if (currentListType !== attrs.list) {
            if (listType) {
              html += `</${listType}>`;
            }
            currentListType = attrs.list;
            listType = attrs.list === 'ordered' ? 'ol' : 'ul';
            const listTypeStyle = attrs.list === 'ordered' ? 'list-decimal' : 'list-disc';
            html += `<${listType} class="list-inside ${listTypeStyle} my-4">`;
          }
          html += `<li class="mb-1">${trimmedContent}</li>`;
          paragraphContent = '';
        }
      } else if (attrs.blockquote) {
        // Close any open list before blockquote
        if (listType) {
          html += `</${listType}>`;
          listType = null;
          currentListType = null;
        }
        const trimmedContent = paragraphContent.trim();
        if (!inBlockquote && trimmedContent) {
          html += `
          <blockquote class="p-4 my-4 border-s-4 border-gray-300 bg-gray-50 dark:border-gray-500 dark:bg-gray-800">
            <p>${trimmedContent}</p>
          </blockquote>`;
          paragraphContent = '';
          inBlockquote = true;
        }
      } else {
        // Regular paragraph break
        if (!nextOp?.attributes?.list) {
          if (listType) {
            html += `</${listType}>`;
            listType = null;
            currentListType = null;
          }
          wrapBlock(true);
        }
        if (inBlockquote) {
          inBlockquote = false;
        }
      }
    } else {
      const formattedText = processInlineFormatting(text, attrs);
      paragraphContent += formattedText;
    }
  });

  // Clean up any remaining content
  const finalContent = paragraphContent.trim();
  if (finalContent) {
    if (listType) {
      html += `<li>${finalContent}</li>`;
    } else {
      html += `<p class="mb-2">${finalContent}</p>`;
    }
  }
  if (listType) html += `</${listType}>`;
  if (inBlockquote && finalContent) html += '</blockquote>';

  return html;
};
