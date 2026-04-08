"use client";

import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { truncateText } from "../../lib/format";

interface TruncatedTextProps {
  text: string;
  maxLength?: number;
  className?: string;
}

export function TruncatedText({ text, maxLength = 200, className = "" }: TruncatedTextProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  const { displayText, isTruncated } = truncateText(text, maxLength);

  if (!isTruncated) {
    return <span className={className}>{text}</span>;
  }

  return (
    <div className={`inline-flex flex-col gap-1 ${className}`}>
      <span className="whitespace-pre-wrap">
        {isExpanded ? text : displayText}
      </span>
      <button
        type="button"
        onClick={() => setIsExpanded(!isExpanded)}
        className="inline-flex items-center gap-1 text-sm text-muted hover:text-ink transition-colors"
      >
        {isExpanded ? (
          <>
            <ChevronUp size={14} />
            Show less
          </>
        ) : (
          <>
            <ChevronDown size={14} />
            Show more
          </>
        )}
      </button>
    </div>
  );
}