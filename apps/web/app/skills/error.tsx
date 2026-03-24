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
        <h2>Failed to load skills</h2>
        <p>Could not reach the API. Check that the backend is running on port 8080.</p>
        <button onClick={reset} className="btn-retry">
          Try again
        </button>
      </div>
    </div>
  );
}
