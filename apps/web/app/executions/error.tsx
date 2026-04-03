"use client";

import { useEffect } from "react";
import { useTranslations } from "next-intl";

export default function ErrorPanel({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  const t = useTranslations("errors");

  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="error-shell">
      <div className="error-panel">
        <h2>Failed to load executions</h2>
        <p>{t("network")}</p>
        <button onClick={reset} className="btn-retry">
          Try again
        </button>
      </div>
    </div>
  );
}