"use client";

import { useState, useRef, useEffect, ReactNode } from "react";

interface TooltipProps {
  content: ReactNode;
  children: ReactNode;
  position?: "right" | "bottom";
  delay?: number;
}

export default function Tooltip({ content, children, position = "right", delay = 150 }: TooltipProps) {
  const [isVisible, setIsVisible] = useState(false);
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);
  const triggerRef = useRef<HTMLDivElement>(null);

  function showTooltip() {
    timeoutRef.current = setTimeout(() => {
      setIsVisible(true);
    }, delay);
  }

  function hideTooltip() {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    setIsVisible(false);
  }

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  const positionClasses = position === "right"
    ? "left-full top-1/2 -translate-y-1/2 ml-2"
    : "top-full left-1/2 -translate-x-1/2 mt-2";

  return (
    <div
      ref={triggerRef}
      className="relative inline-flex"
      onMouseEnter={showTooltip}
      onMouseLeave={hideTooltip}
    >
      {children}
      {isVisible && (
        <div
          className={`absolute z-50 whitespace-nowrap px-3 py-2 text-sm rounded-md shadow-lg border border-line bg-surface-raised text-ink pointer-events-none animate-in fade-in zoom-in-95 duration-150 ${positionClasses}`}
          style={{
            backdropFilter: "blur(8px)",
          }}
        >
          {content}
          <div
            className={`absolute w-2 h-2 bg-surface-raised border-line rotate-45 ${position === "right" ? "-left-1 top-1/2 -translate-y-1/2 border-r-0 border-t-0" : "-top-1 left-1/2 -translate-x-1/2 border-b-0 border-l-0"}`}
            style={{
              backgroundColor: "var(--surface-raised)",
              borderColor: "var(--line)",
            }}
          />
        </div>
      )}
    </div>
  );
}
