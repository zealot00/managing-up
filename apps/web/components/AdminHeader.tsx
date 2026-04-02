"use client";

import { usePathname } from "next/navigation";
import { useAuth } from "../context/AuthContext";

const pageTitles: Record<string, string> = {
  "/dashboard": "Dashboard",
  "/skills": "Skills Registry",
  "/executions": "Executions",
  "/approvals": "Approvals",
  "/tasks": "Tasks",
  "/evaluations": "Evaluations",
  "/experiments": "Experiments",
  "/replays": "Replays",
  "/gateway": "Gateway",
  "/seh": "SEH Module",
};

export default function AdminHeader() {
  const pathname = usePathname();
  const { user, isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return null;
  }

  const pageTitle = Object.entries(pageTitles).find(
    ([path]) => pathname === path || pathname.startsWith(path + "/")
  )?.[1] || "Managing Up";

  return (
    <header className="admin-header">
      <div className="admin-header-left">
        <h1 className="admin-header-title">{pageTitle}</h1>
      </div>
      <div className="admin-header-right">
        <div className="admin-header-search">
          <span className="admin-header-search-icon">⌕</span>
          <input
            type="text"
            placeholder="Search..."
            className="admin-header-search-input"
          />
          <kbd className="admin-header-search-kbd">⌘K</kbd>
        </div>
        <button className="admin-header-icon-btn" title="Notifications">
          <span>🔔</span>
          <span className="admin-header-badge">3</span>
        </button>
        <button className="admin-header-icon-btn" title="Help">
          <span>❓</span>
        </button>
      </div>
    </header>
  );
}
