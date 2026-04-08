import { formatDistanceToNow, format, formatDuration, intervalToDuration, type Locale } from "date-fns";
import { zhCN, enUS } from "date-fns/locale";

const locales: Record<string, Locale> = {
  en: enUS,
  zh: zhCN,
};

// Get locale from next-intl or default to en
function getLocale(): Locale {
  // Try to get locale from window or default
  if (typeof window !== "undefined") {
    const htmlLang = document.documentElement.lang || "en";
    return locales[htmlLang] || enUS;
  }
  return enUS;
}

/**
 * Format a date as relative time (e.g., "2 mins ago", "Just now")
 * For dates within the last 24 hours, shows relative time
 * For older dates, shows formatted date
 */
export function formatRelativeTime(date: string | Date): string {
  const d = typeof date === "string" ? new Date(date) : date;

  if (isNaN(d.getTime())) return "—";

  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffHours = diffMs / (1000 * 60 * 60);

  // Within 24 hours: show relative time
  if (diffHours < 24) {
    return formatDistanceToNow(d, { addSuffix: true, locale: getLocale() });
  }

  // Older: show formatted date
  return format(d, "MMM d, yyyy", { locale: getLocale() });
}

/**
 * Format a date as full datetime
 */
export function formatDateTime(date: string | Date): string {
  const d = typeof date === "string" ? new Date(date) : date;
  if (isNaN(d.getTime())) return "—";
  return format(d, "MMM d, yyyy HH:mm", { locale: getLocale() });
}

/**
 * Format a date for display in lists
 */
export function formatDate(date: string | Date): string {
  const d = typeof date === "string" ? new Date(date) : date;
  if (isNaN(d.getTime())) return "—";
  return format(d, "MMM d, yyyy");
}

/**
 * Format duration in milliseconds to human-readable string
 * Examples: 125430 -> "2m 5s", 500 -> "0.5s", 65000 -> "1m 5s"
 */
export function formatDurationMs(ms: number): string {
  if (ms < 0) return "—";
  if (ms < 1000) return `${ms}ms`;

  if (ms < 60000) {
    // Less than 1 minute: show seconds with 1 decimal
    const seconds = ms / 1000;
    return `${seconds.toFixed(1)}s`;
  }

  // 1 minute or more: show minutes and seconds
  const duration = intervalToDuration({ start: 0, end: ms });
  return formatDuration(duration, { locale: getLocale() });
}

/**
 * Format duration in seconds (common API response format)
 */
export function formatDurationSeconds(seconds: number): string {
  return formatDurationMs(seconds * 1000);
}

/**
 * Format a number as percentage
 */
export function formatPercent(value: number, decimals = 1): string {
  if (isNaN(value) || value < 0) return "—";
  return `${(value * 100).toFixed(decimals)}%`;
}

/**
 * Format a score with color indicator
 */
export function formatScore(score: number): string {
  const percent = score * 100;
  return `${percent.toFixed(1)}%`;
}

/**
 * Truncate long text with expand/collapse
 * Returns { text, isTruncated, displayText }
 */
export function truncateText(text: string, maxLength: number = 200): {
  displayText: string;
  isTruncated: boolean;
} {
  if (!text || text.length <= maxLength) {
    return { displayText: text, isTruncated: false };
  }

  return {
    displayText: text.slice(0, maxLength).trimEnd() + "...",
    isTruncated: true,
  };
}

/**
 * Format bytes to human-readable string
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}