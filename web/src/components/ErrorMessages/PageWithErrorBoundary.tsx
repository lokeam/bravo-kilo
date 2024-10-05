import React from 'react';
import { ReactNode } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import ErrorFallbackPage from '../../pages/ErrorFallback';

interface PageWithErrorBoundaryProps {
  children: ReactNode;
  fallbackMessage: string;
}

const PageWithErrorBoundary: React.FC<PageWithErrorBoundaryProps> = ({ children, fallbackMessage }) => {
  return (
    <ErrorBoundary FallbackComponent={(props) => <ErrorFallbackPage {...props} message={fallbackMessage} />}>
      {children}
    </ErrorBoundary>
  );
}

export default PageWithErrorBoundary;
