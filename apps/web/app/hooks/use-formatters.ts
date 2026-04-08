"use client";

import { useLocale } from "next-intl";
import {
  formatRelativeTime,
  formatDateTime,
  formatDate,
  formatDurationMs,
  formatDurationSeconds,
  formatPercent,
  formatScore,
  truncateText,
  formatBytes,
} from "../lib/format";

export function useFormatters() {
  const locale = useLocale();

  return {
    relativeTime: formatRelativeTime,
    dateTime: formatDateTime,
    date: formatDate,
    durationMs: formatDurationMs,
    durationSeconds: formatDurationSeconds,
    percent: formatPercent,
    score: formatScore,
    truncate: truncateText,
    bytes: formatBytes,
  };
}