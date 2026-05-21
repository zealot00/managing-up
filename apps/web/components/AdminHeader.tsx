"use client";

import { useState, useEffect, useCallback } from "react";
import { useAuth } from "../context/AuthContext";
import { Bell, CircleHelp, Search, Menu, Command } from "lucide-react";
import { useMobileSidebar } from "./MobileSidebarProvider";
import { CommandPalette } from "../app/components/ui/CommandPalette";
import LanguageSwitcher from "./LanguageSwitcher";

export default function AdminHeader() {
  const { isAuthenticated } = useAuth();
  const { toggle } = useMobileSidebar();
  const [isSearchOpen, setIsSearchOpen] = useState(false);

  // Cmd+K / Ctrl+K global shortcut
  const openSearch = useCallback(() => setIsSearchOpen(true), []);
  const closeSearch = useCallback(() => setIsSearchOpen(false), []);

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setIsSearchOpen((prev) => !prev);
      }
    }
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  if (!isAuthenticated) {
    return null;
  }

  return (
    <>
      <header className="admin-header">
        <div className="admin-header-left">
          <button
            className="admin-header-menu-btn"
            onClick={toggle}
            aria-label="Open menu"
          >
            <Menu size={20} aria-hidden="true" />
          </button>
        </div>
        <div className="admin-header-right">
          <button className="admin-header-search" onClick={openSearch} type="button" aria-label="Search">
            <Search size={16} className="admin-header-search-icon" aria-hidden="true" />
            <span className="admin-header-search-input">Search...</span>
            <kbd className="admin-header-search-kbd">
              <Command size={12} aria-hidden="true" />
              <span>K</span>
            </kbd>
          </button>
          <button className="admin-header-icon-btn" title="Notifications" aria-label="Notifications">
            <Bell size={20} aria-hidden="true" />
          </button>
          <button className="admin-header-icon-btn" title="Help" aria-label="Help">
            <CircleHelp size={20} aria-hidden="true" />
          </button>
          <div className="admin-header-lang" title="Switch language">
            <LanguageSwitcher />
          </div>
        </div>
      </header>
      <CommandPalette isOpen={isSearchOpen} onClose={closeSearch} />
    </>
  );
}
