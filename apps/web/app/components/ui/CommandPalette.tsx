"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { Search, ArrowRight } from "lucide-react";
import { useFocusTrap } from "../../hooks/use-focus-trap";

interface NavItem {
  href: string;
  label: string;
  group: string;
  keywords: string[];
}

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
}

export function CommandPalette({ isOpen, onClose }: CommandPaletteProps) {
  const t = useTranslations("commandPalette");
  const tNav = useTranslations("nav");
  const router = useRouter();
  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);
  const listRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const trapRef = useFocusTrap(isOpen);

  const navItems: NavItem[] = [
    { href: "/dashboard", label: tNav("dashboard"), group: t("groupNavigation"), keywords: ["home", "overview"] },
    { href: "/skills", label: tNav("skills"), group: t("groupNavigation"), keywords: ["skill", "registry"] },
    { href: "/executions", label: tNav("executions"), group: t("groupNavigation"), keywords: ["execution", "run"] },
    { href: "/tasks", label: tNav("tasks"), group: t("groupNavigation"), keywords: ["task", "eval"] },
    { href: "/tasks/from-trace", label: tNav("taskBuilder"), group: t("groupNavigation"), keywords: ["builder", "trace"] },
    { href: "/evaluations", label: tNav("evaluations"), group: t("groupNavigation"), keywords: ["evaluation", "metric"] },
    { href: "/experiments", label: tNav("experiments"), group: t("groupNavigation"), keywords: ["experiment", "compare"] },
    { href: "/approvals", label: tNav("approvals"), group: t("groupNavigation"), keywords: ["approval", "review"] },
    { href: "/replays", label: tNav("replays"), group: t("groupNavigation"), keywords: ["replay", "snapshot"] },
    { href: "/gateway", label: tNav("gateway"), group: t("groupNavigation"), keywords: ["gateway", "key", "api"] },
    { href: "/gateway/providers", label: tNav("providers"), group: t("groupNavigation"), keywords: ["provider", "key"] },
    { href: "/mcp", label: tNav("mcp"), group: t("groupNavigation"), keywords: ["mcp", "server", "tool"] },
    { href: "/mcp-router", label: tNav("mcpRouter"), group: t("groupNavigation"), keywords: ["mcp", "router", "route"] },
    { href: "/sweeps", label: tNav("sweeps"), group: t("groupNavigation"), keywords: ["sweep", "hyperparameter"] },
    { href: "/policies", label: tNav("policies"), group: t("groupNavigation"), keywords: ["policy", "governance"] },
    { href: "/fallback-chains", label: tNav("fallbackChains"), group: t("groupNavigation"), keywords: ["fallback", "chain"] },
    { href: "/seh/datasets", label: tNav("sehDatasets"), group: t("groupNavigation"), keywords: ["seh", "dataset"] },
    { href: "/seh/runs", label: tNav("sehRuns"), group: t("groupNavigation"), keywords: ["seh", "run"] },
    { href: "/seh/policies", label: tNav("sehPolicies"), group: t("groupNavigation"), keywords: ["seh", "policy"] },
    { href: "/profile", label: tNav("profile"), group: t("groupSettings"), keywords: ["profile", "account", "password"] },
    { href: "/preferences", label: tNav("preferences"), group: t("groupSettings"), keywords: ["preferences", "settings", "language"] },
  ];

  const filtered = query
    ? navItems.filter((item) => {
        const q = query.toLowerCase();
        return (
          item.label.toLowerCase().includes(q) ||
          item.href.toLowerCase().includes(q) ||
          item.keywords.some((kw) => kw.includes(q))
        );
      })
    : navItems;

  // Reset state when opened/closed
  useEffect(() => {
    if (isOpen) {
      setQuery("");
      setActiveIndex(0);
      // Focus the input after the dialog renders
      requestAnimationFrame(() => {
        inputRef.current?.focus();
      });
    }
  }, [isOpen]);

  // Reset active index when query changes
  useEffect(() => {
    setActiveIndex(0);
  }, [query]);

  const handleSelect = useCallback(
    (item: NavItem) => {
      onClose();
      router.push(item.href);
    },
    [onClose, router],
  );

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setActiveIndex((prev) => (prev + 1) % filtered.length);
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setActiveIndex((prev) => (prev - 1 + filtered.length) % filtered.length);
      } else if (e.key === "Enter") {
        e.preventDefault();
        if (filtered[activeIndex]) {
          handleSelect(filtered[activeIndex]);
        }
      } else if (e.key === "Escape") {
        e.preventDefault();
        onClose();
      }
    },
    [filtered, activeIndex, handleSelect, onClose],
  );

  // Scroll active item into view
  useEffect(() => {
    if (!listRef.current) return;
    const activeEl = listRef.current.querySelector('[data-command-active="true"]');
    activeEl?.scrollIntoView({ block: "nearest" });
  }, [activeIndex]);

  if (!isOpen) return null;

  const titleId = "command-palette-title";

  return (
    <div className="command-palette-backdrop" onClick={onClose} role="presentation">
      <div
        ref={trapRef}
        className="command-palette"
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        <div className="command-palette-header">
          <Search size={18} className="command-palette-search-icon" aria-hidden="true" />
          <input
            ref={inputRef}
            type="text"
            className="command-palette-input"
            placeholder={t("placeholder")}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            aria-label={t("placeholder")}
          />
        </div>
        <div className="command-palette-list" ref={listRef} role="listbox">
          {filtered.length === 0 ? (
            <div className="command-palette-empty">{t("noResults")}</div>
          ) : (
            (() => {
              let currentGroup = "";
              const elements: React.ReactNode[] = [];
              filtered.forEach((item, index) => {
                if (item.group !== currentGroup) {
                  currentGroup = item.group;
                  elements.push(
                    <div key={`group-${currentGroup}`} className="command-palette-group" role="presentation">
                      {currentGroup}
                    </div>,
                  );
                }
                elements.push(
                  <button
                    key={item.href}
                    className={`command-palette-item${index === activeIndex ? " is-active" : ""}`}
                    onClick={() => handleSelect(item)}
                    onMouseEnter={() => setActiveIndex(index)}
                    role="option"
                    aria-selected={index === activeIndex}
                    data-command-active={index === activeIndex ? "true" : undefined}
                  >
                    <span className="command-palette-item-label">{item.label}</span>
                    <span className="command-palette-item-path">{item.href}</span>
                    <ArrowRight size={14} className="command-palette-item-arrow" aria-hidden="true" />
                  </button>,
                );
              });
              return elements;
            })()
          )}
        </div>
        <div className="command-palette-footer">
          <kbd>↑↓</kbd> <span>{t("navigate")}</span>
          <kbd>↵</kbd> <span>{t("select")}</span>
          <kbd>esc</kbd> <span>{t("close")}</span>
        </div>
      </div>
    </div>
  );
}
