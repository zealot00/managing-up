"use client";

import { useState, useEffect } from "react";

/**
 * Debounces a value by the given delay in milliseconds.
 * Returns the debounced value that only updates after the delay elapses
 * with no new changes.
 */
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => clearTimeout(timer);
  }, [value, delay]);

  return debouncedValue;
}
