"use client";

import { useEffect } from "react";

export default function ErrorPanel({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="error-shell">
      <div className="error-panel">
        <h2>Something went wrong</h2>
        <p>The page failed to load. The backend may be unreachable.</p>
        <button onClick={reset} className="btn-retry">
          Try again
        </button>
      </div>
    </div>
  );
}
