import { FallbackProps } from 'react-error-boundary';

interface ErrorFallbackPageProps extends FallbackProps {
  message?: string;
}

function ErrorFallbackPage({ error, resetErrorBoundary, message }: ErrorFallbackPageProps) {
  if (error) {
    console.log(`page error | message - ${message} | error - ${error}`);
  }

  return (
    <section className="error_fallback bg-white dark:bg-black min-h-screen bg-cover flex flex-col items-center antialiased mdTablet:pl-1 pr-5 pt-36 mdTablet:ml-24">
      <div
        className="mx-auto max-w-screen-sm text-center"
        role="alert"
      >

        <h2 className="mb-4 text-7xl tracking-tight font-extrabold lg:text-9xl text-primary-600 dark:text-primary-500">Oops! Something went wrong</h2>
        <p className="mb-4 text-3xl tracking-tight font-bold text-gray-900 md:text-4xl dark:text-white">We apologize for any inconvenience.</p>
        <p className="mb-4 text-lg font-light text-charcoal dark:text-white-smoke">We're working to get things back up and running. Please try again later.</p>
        <button
          className="bg-vivid-blue hover:bg-vivid-blue-d hover:border-vivid-blue-d dark:bg-vivid-blue dark:hover:bg-vivid-blue-d dark:hover:border-vivid-blue-d dark:hover:text-white transition duration-500 ease-in-out font-bold"
          onClick={resetErrorBoundary}
        >
          Try again
        </button>
      </div>
    </section>
  );
}

export default ErrorFallbackPage;
