"use client";

import { usePathname } from "next/navigation";
import { useAuth } from "../context/AuthContext";
import { useTranslations } from "next-intl";
import { Bell, CircleHelp, Search, Menu, Command, Globe } from "lucide-react";
import { useMobileSidebar } from "./MobileSidebarProvider";
import LanguageSwitcher from "./LanguageSwitcher";

export default function AdminHeader() {
  const t = useTranslations("nav");
  const pathname = usePathname();
  const { user, isAuthenticated } = useAuth();
  const { toggle } = useMobileSidebar();

  const pageTitles: Record<string, string> = {
    "/dashboard": t("dashboard"),
    "/skills": t("skills"),
    "/executions": t("executions"),
    "/approvals": t("approvals"),
    "/tasks": t("tasks"),
    "/evaluations": t("evaluations"),
    "/experiments": t("experiments"),
    "/replays": t("replays"),
    "/gateway": t("gateway"),
    "/seh": t("sehModule"),
    "/sweeps": t("sweeps"),
  };

  if (!isAuthenticated) {
    return null;
  }

  const pageTitle = Object.entries(pageTitles).find(
    ([path]) => pathname === path || pathname.startsWith(path + "/")
  )?.[1] || "Managing Up";

  return (
    <header className="admin-header">
      <div className="admin-header-left">
        <button
          className="admin-header-menu-btn"
          onClick={toggle}
          aria-label="Open menu"
        >
          <Menu size={20} aria-hidden="true" />
        </button>
        <h1 className="admin-header-title">{pageTitle}</h1>
      </div>
      <div className="admin-header-right">
        <div className="admin-header-search">
          <Search size={16} className="admin-header-search-icon" aria-hidden="true" />
          <input
            type="text"
            placeholder="Search..."
            className="admin-header-search-input"
          />
          <kbd className="admin-header-search-kbd">
            <Command size={12} aria-hidden="true" />
            <span>K</span>
          </kbd>
        </div>
        <button className="admin-header-icon-btn" title="Notifications" aria-label="Notifications">
          <Bell size={20} aria-hidden="true" />
          <span className="admin-header-badge">3</span>
        </button>
        <button className="admin-header-icon-btn" title="Help" aria-label="Help">
          <CircleHelp size={20} aria-hidden="true" />
        </button>
        <div className="admin-header-lang" title="Switch language">
          <LanguageSwitcher />
        </div>
      </div>
    </header>
  );
}
