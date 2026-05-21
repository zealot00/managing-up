"use client";

import { useEffect, useRef, useCallback } from "react";

const FOCUSABLE_SELECTOR = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
  '[contenteditable]',
].join(", ");

function getFocusableElements(container: HTMLElement): HTMLElement[] {
  return Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR));
}

/**
 * Focus trap hook for modal/dialog/drawer components.
 *
 * - Traps Tab / Shift+Tab within the container element
 * - Auto-focuses the first focusable element on mount
 * - Restores focus to the previously active element on unmount
 */
export function useFocusTrap(isActive: boolean) {
  const containerRef = useRef<HTMLElement | null>(null);
  const previousActiveRef = useRef<HTMLElement | null>(null);

  const setRef = useCallback(
    (node: HTMLElement | null) => {
      containerRef.current = node;
    },
    [],
  );

  useEffect(() => {
    if (!isActive) return;

    // Save the currently focused element so we can restore it later
    previousActiveRef.current = document.activeElement as HTMLElement;

    const container = containerRef.current;
    if (!container) return;

    // Auto-focus the first focusable element
    const focusables = getFocusableElements(container);
    if (focusables.length > 0) {
      // Slight delay to allow animations to start
      const timerId = requestAnimationFrame(() => {
        focusables[0].focus();
      });
    }

    function handleKeyDown(e: KeyboardEvent) {
      if (e.key !== "Tab") return;

      const els = getFocusableElements(container!);
      if (els.length === 0) {
        e.preventDefault();
        return;
      }

      const first = els[0];
      const last = els[els.length - 1];

      if (e.shiftKey) {
        if (document.activeElement === first) {
          e.preventDefault();
          last.focus();
        }
      } else {
        if (document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    }

    container.addEventListener("keydown", handleKeyDown);

    return () => {
      container.removeEventListener("keydown", handleKeyDown);
      // Restore focus to the previously active element
      if (previousActiveRef.current && typeof previousActiveRef.current.focus === "function") {
        previousActiveRef.current.focus();
      }
      previousActiveRef.current = null;
    };
  }, [isActive]);

  return setRef;
}
