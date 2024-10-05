import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import userEvent from '@testing-library/user-event';
import { test, expect, afterEach, beforeEach, describe, vi,  } from 'vitest';
import useAddBook from '../../../../src/hooks/useAddBook';
import ManualAddBook from '../../../../src/pages/ManualAddBook';
import '@testing-library/jest-dom';

vi.mock('react-router-dom', async (importOriginal) => {
  const actual = await importOriginal();
  return {
    ...(actual as object),
    useNavigate: vi.fn(),
    useLocation: vi.fn().mockReturnValue({ state: { book: null } }),
  };
});

vi.mock('../../../../src/hooks/useAddBook');

describe('ManualAddBook Component', () => {
  beforeEach(() => {
    vi.mocked(useAddBook).mockReturnValue({ mutate: vi.fn() });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  // Ensure checkboxes behave properly
  test('allows user to check and uncheck formats', () => {
    render(
      <MemoryRouter>
        <ManualAddBook />
      </MemoryRouter>
    );

    const physicalLabel = screen.getByLabelText(/physical/i);
    const ebookLabel = screen.getByLabelText(/ebook/i);
    const audiobookLabel = screen.getByLabelText(/audiobook/i);

    // Simulate checking the checkboxes by clicking the labels
    fireEvent.click(physicalLabel);
    fireEvent.click(ebookLabel);
    fireEvent.click(audiobookLabel);

    // Ensure the checkboxes are checked
    const physicalCheckbox = screen.getByRole('checkbox', { name: /physical/i });
    const ebookCheckbox = screen.getByRole('checkbox', { name: /ebook/i });
    const audiobookCheckbox = screen.getByRole('checkbox', { name: /audiobook/i });

    expect(physicalCheckbox).toBeChecked();
    expect(ebookCheckbox).toBeChecked();
    expect(audiobookCheckbox).toBeChecked();

    // Simulate unchecking by clicking the labels again
    fireEvent.click(physicalLabel);
    fireEvent.click(ebookLabel);
    fireEvent.click(audiobookLabel);

    // Ensure the checkboxes are unchecked
    expect(physicalCheckbox).not.toBeChecked();
    expect(ebookCheckbox).not.toBeChecked();
    expect(audiobookCheckbox).not.toBeChecked();
  });

  // Test field interactions and validation
  test('handles form input and displays validation errors correctly', async () => {
    render(
      <MemoryRouter>
        <ManualAddBook />
      </MemoryRouter>
    );

    const titleInput = screen.getByRole('textbox', { name: /title \*/i });
    const isbn10Input = screen.getByRole('textbox', { name: /isbn-10/i });
    const isbn13Input = screen.getByRole('textbox', { name: /isbn-13/i });
    const languageInput = screen.getByRole('textbox', { name: /language/i });

    // Fill in valid inputs
    fireEvent.change(titleInput, { target: { value: 'My Book Title' } });
    fireEvent.change(isbn10Input, { target: { value: '1234567890' } });
    fireEvent.change(isbn13Input, { target: { value: '1234567890123' } });
    fireEvent.change(languageInput, { target: { value: 'English' } });

    expect(titleInput).toHaveValue('My Book Title');
    expect(isbn10Input).toHaveValue('1234567890');
    expect(isbn13Input).toHaveValue('1234567890123');
    expect(languageInput).toHaveValue('English');

    // Simulate form submission with valid data
    const submitButton = screen.getByRole('button', { name: /add book/i });
    fireEvent.click(submitButton);

    // Ensure no validation errors for valid input
    await vi.waitFor(() => {
      expect(screen.queryByText(/please enter a title/i)).not.toBeInTheDocument();
      expect(screen.queryByText(/isbn10 must contain 10 characters/i)).not.toBeInTheDocument();
      expect(screen.queryByText(/isbn13 must contain 13 characters/i)).not.toBeInTheDocument();
      expect(screen.queryByText(/please enter a language/i)).not.toBeInTheDocument();
    });
  });

  // Test validation
  test('shows validation errors for missing or invalid data', async () => {
    render(<ManualAddBook />);

    console.log('Before clicking submit'); // This will log before the click event

    const submitButton = screen.getByRole('button', { name: /add book/i });
    userEvent.click(submitButton);

    console.log('After clicking submit'); // This logs after submitting

    screen.debug(); // This will print the DOM tree
  });

  // Test field arrays
  test('handles adding and removing authors correctly', async () => {
    render(
      <MemoryRouter>
        <ManualAddBook />
      </MemoryRouter>
    );

    // No initial author input fields
    expect(screen.queryByTestId('author-input-0')).not.toBeInTheDocument();

    // Click 'Add Another Author' to add a new input field for authors
    const addAuthorButton = screen.getByText(/add another author/i);
    fireEvent.click(addAuthorButton);

    // Check if first author input is added
    const firstAuthorInput = screen.getByTestId('author-input-0');
    expect(firstAuthorInput).toBeInTheDocument();

    // Type into first author input
    fireEvent.change(firstAuthorInput, { target: { value: 'First Author' } });
    expect(firstAuthorInput).toHaveValue('First Author');

    // Click 'Add Another Author' to add second input field
    fireEvent.click(addAuthorButton);

    // Check if second author input is added
    const secondAuthorInput = screen.getByTestId('author-input-1');
    expect(secondAuthorInput).toBeInTheDocument();

    // Type into second author input
    fireEvent.change(secondAuthorInput, { target: { value: 'Second Author' } });
    expect(secondAuthorInput).toHaveValue('Second Author');

    // Click the remove button for the second author
    const removeButton = screen.getByLabelText('Remove Author 2'); // Match the aria-label
    fireEvent.click(removeButton);

    // Ensure the second author input is removed
    expect(secondAuthorInput).not.toBeInTheDocument();
  });

});
