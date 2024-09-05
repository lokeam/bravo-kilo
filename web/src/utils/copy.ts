import DOMPurify from 'dompurify';

export const copyTextContent = () => {
  const element = document.getElementById("prompt_response");

  if (element) {
    // Get the text content with appropriate handling for line breaks
    const textContent = Array.from(element.childNodes)
      .map(node => (node.textContent || '').trim())
      .join('\n\n'); // Use double newline for clearer formatting

    // Sanitize the content using DOMPurify
    const sanitizedText = DOMPurify.sanitize(textContent);

    // Copy the sanitized content using the Clipboard API
    navigator.clipboard.writeText(sanitizedText).then(
      () => alert("Text copied to clipboard!"),
      (err) => console.error("Failed to copy text: ", err)
    );
  }
};
