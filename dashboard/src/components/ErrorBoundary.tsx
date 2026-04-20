'use client';

import React, { ReactNode, ErrorInfo } from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: (error: Error, reset: () => void) => ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
  errorCount: number;
}

/**
 * ErrorBoundary Component
 * Catches React component errors and displays a user-friendly error UI
 * with recovery options and detailed error logging for debugging
 */
export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  private resetTimeout: NodeJS.Timeout | null = null;

  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorCount: 0,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return {
      hasError: true,
      error,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Log error details for debugging
    console.error('ErrorBoundary caught an error:', error);
    console.error('Component stack:', errorInfo.componentStack);

    // Update state with error details
    this.setState((prevState) => ({
      errorInfo,
      errorCount: prevState.errorCount + 1,
    }));

    // Send to error tracking service (e.g., Sentry) in production
    if (process.env.NODE_ENV === 'production') {
      // Example: Sentry.captureException(error, { contexts: { react: errorInfo } });
      console.error('Sending error to monitoring service:', error.message);
    }

    // Auto-reset after 5 errors within a session (prevents infinite loops)
    if (this.state.errorCount > 5) {
      console.warn('Too many errors detected. Consider reloading the page.');
    }
  }

  handleReset = () => {
    // Clear any pending timeouts
    if (this.resetTimeout) {
      clearTimeout(this.resetTimeout);
    }

    // Reset error state
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });

    // Hard refresh to ensure clean state
    this.resetTimeout = setTimeout(() => {
      window.location.href = '/';
    }, 500);
  };

  componentWillUnmount() {
    if (this.resetTimeout) {
      clearTimeout(this.resetTimeout);
    }
  }

  render() {
    if (this.state.hasError) {
      // Use custom fallback if provided
      if (this.props.fallback && this.state.error) {
        return this.props.fallback(this.state.error, this.handleReset);
      }

      // Default error UI
      return (
        <div className="min-h-screen bg-[#0A0A0A] flex items-center justify-center p-4">
          <div className="max-w-md w-full">
            {/* Error Container */}
            <div className="bg-[#0F0F11] border border-red-500/20 rounded-xl p-8 space-y-6">
              {/* Error Icon */}
              <div className="flex justify-center">
                <div className="bg-red-500/10 p-4 rounded-full">
                  <AlertTriangle className="w-8 h-8 text-red-400" />
                </div>
              </div>

              {/* Error Title */}
              <div className="text-center space-y-2">
                <h1 className="text-2xl font-bold text-white">Something went wrong</h1>
                <p className="text-sm text-slate-400">
                  An unexpected error occurred in the dashboard
                </p>
              </div>

              {/* Error Details (Dev Mode) */}
              {process.env.NODE_ENV === 'development' && this.state.error && (
                <div className="space-y-2">
                  <details className="text-xs">
                    <summary className="cursor-pointer font-semibold text-slate-300 hover:text-slate-200">
                      Error Details
                    </summary>
                    <div className="mt-3 p-3 bg-[#141416] rounded border border-[#27272A] space-y-2">
                      <div>
                        <span className="text-slate-500">Message:</span>
                        <p className="text-red-400 font-mono mt-1">
                          {this.state.error.message}
                        </p>
                      </div>
                      {this.state.errorInfo?.componentStack && (
                        <div>
                          <span className="text-slate-500">Component Stack:</span>
                          <pre className="text-slate-400 font-mono mt-1 text-[10px] overflow-auto max-h-40">
                            {this.state.errorInfo.componentStack}
                          </pre>
                        </div>
                      )}
                    </div>
                  </details>
                </div>
              )}

              {/* Error Counter */}
              {this.state.errorCount > 1 && (
                <div className="text-center">
                  <p className="text-xs text-slate-500">
                    Error count in this session: {this.state.errorCount}
                  </p>
                </div>
              )}

              {/* Action Buttons */}
              <div className="flex flex-col gap-3">
                <button
                  onClick={this.handleReset}
                  className="flex items-center justify-center gap-2 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white font-medium rounded-lg transition-colors"
                >
                  <RefreshCw className="w-4 h-4" />
                  Try Again
                </button>
                <button
                  onClick={() => (window.location.href = '/')}
                  className="px-4 py-2.5 bg-[#1A1A1E] hover:bg-[#27272A] text-slate-300 font-medium rounded-lg border border-[#27272A] transition-colors"
                >
                  Go to Home
                </button>
              </div>

              {/* Support Info */}
              <div className="text-center pt-4 border-t border-[#27272A]">
                <p className="text-xs text-slate-500">
                  If this persists, please contact{' '}
                  <a
                    href="mailto:support@tribunal.dev"
                    className="text-indigo-400 hover:underline"
                  >
                    support
                  </a>
                </p>
              </div>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
