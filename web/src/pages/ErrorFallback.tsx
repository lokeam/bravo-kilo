import { FallbackProps } from 'react-error-boundary';
import UnplugIcon from '../components/CustomSVGs/UnplugIcon';

interface ErrorFallbackPageProps extends FallbackProps {
  message?: string;
}

function ErrorFallbackPage({ error, resetErrorBoundary, message }: ErrorFallbackPageProps) {
  if (error) {
    console.log(`page error | message - ${message} | error - ${error}`);
  }

  return (
    <section>
      <div className="px-6 py-16 mx-auto max-w-screen-mdDesktop"></div>
      <div
        className="error_fallback bg-white dark:bg-gray-900 text-center max-w-screen-lgMobile mx-auto"
        role="alert"
      >
        <UnplugIcon />
        <h2>Oops! Something went wrong</h2>
        <p>We apologize for any inconvenience.'</p>
        <p>We're working to get things back up and running. Please try again later.</p>
        <button onClick={resetErrorBoundary}>Try again</button>
      </div>
    </section>
  );
}

export default ErrorFallbackPage;
