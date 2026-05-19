"use client";

import { useEffect } from "react";
import { AlertTriangle, RefreshCw } from "lucide-react";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Application error:", error);
  }, [error]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--bg)]">
      <div className="text-center max-w-md mx-auto px-4">
        <div className="inline-flex items-center justify-center w-16 h-16 bg-[var(--danger-bg)] rounded-full mb-6">
          <AlertTriangle className="w-8 h-8 text-[var(--danger)]" />
        </div>
        <h1 className="text-2xl font-bold text-[var(--ink-strong)]">
          Something went wrong
        </h1>
        <p className="text-[var(--muted)] mt-2">
          We have been notified and are working to fix the issue.
        </p>
        {error.digest && (
          <p className="text-xs text-[var(--muted)] mt-4 font-mono">
            Error ID: {error.digest}
          </p>
        )}
        <button
          onClick={reset}
          className="inline-flex items-center gap-2 px-4 py-2 bg-[var(--primary)] text-white rounded-md hover:bg-[var(--primary-deep)] transition-colors cursor-pointer mt-6"
        >
          <RefreshCw size={16} />
          Try Again
        </button>
      </div>
    </div>
  );
}