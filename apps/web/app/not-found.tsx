"use client";

import Link from "next/link";
import { Home, ArrowLeft } from "lucide-react";

export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--bg)]">
      <div className="text-center">
        <h1 className="text-6xl font-bold text-[var(--ink-strong)]">404</h1>
        <h2 className="text-2xl font-semibold text-[var(--ink)] mt-4">
          Page not found
        </h2>
        <p className="text-[var(--muted)] mt-2">
          The page you are looking for does not exist or has been moved.
        </p>
        <div className="flex gap-4 justify-center mt-8">
          <Link
            href="/"
            className="inline-flex items-center gap-2 px-4 py-2 bg-[var(--primary)] text-white rounded-md hover:bg-[var(--primary-deep)] transition-colors cursor-pointer"
          >
            <Home size={16} />
            Go Home
          </Link>
          <button
            onClick={() => window.history.back()}
            className="inline-flex items-center gap-2 px-4 py-2 border border-[var(--line-strong)] rounded-md hover:bg-[var(--bg-deep)] transition-colors cursor-pointer"
          >
            <ArrowLeft size={16} />
            Go Back
          </button>
        </div>
      </div>
    </div>
  );
}