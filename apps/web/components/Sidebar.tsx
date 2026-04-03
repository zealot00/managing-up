"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "../context/AuthContext";
import { useRouter } from "next/navigation";
import { useState } from "react";

interface NavItem {
  href: string;
  label: string;
  icon: string;
  children?: { href: string; label: string }[];
}

const navSections: { title: string; items: NavItem[] }[] = [
  {
    title: "Overview",
    items: [
      { href: "/dashboard", label: "Dashboard", icon: "◉" },
    ],
  },
  {
    title: "AI Quality",
    items: [
      { href: "/skills", label: "Skills", icon: "◈" },
      { href: "/executions", label: "Executions", icon: "▸" },
      { href: "/tasks", label: "Tasks", icon: "◉", children: [{ href: "/tasks/from-trace", label: "Task Builder" }] },
      { href: "/evaluations", label: "Evaluations", icon: "◎" },
      { href: "/experiments", label: "Experiments", icon: "◇" },
    ],
  },
  {
    title: "Operations",
    items: [
      { href: "/approvals", label: "Approvals", icon: "◐" },
      { href: "/replays", label: "Replays", icon: "↻" },
      { href: "/gateway", label: "Gateway", icon: "⬡" },
    ],
  },
  {
    title: "System",
    items: [
      { href: "/seh", label: "SEH Module", icon: "▣" },
    ],
  },
];

export default function Sidebar() {
  const pathname = usePathname();
  const { user, isAuthenticated, isLoading, logout } = useAuth();
  const router = useRouter();
  const [collapsed, setCollapsed] = useState(false);

  async function handleLogout() {
    await logout();
    router.push("/login");
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <aside className={`sidebar ${collapsed ? "sidebar-collapsed" : ""}`}>
      <div className="sidebar-header">
        <a href="/" className="sidebar-brand">
          <img src="/logo.svg" alt="managing-up logo" className="sidebar-logo" />
          {!collapsed && (
            <div className="sidebar-brand-text">
              <span className="sidebar-brand-name">managing-up</span>
              <span className="sidebar-brand-edition">EE</span>
            </div>
          )}
        </a>
        <button
          className="sidebar-toggle"
          onClick={() => setCollapsed(!collapsed)}
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        >
          {collapsed ? "▸" : "◂"}
        </button>
      </div>

      <nav className="sidebar-nav">
        {navSections.map((section) => (
          <div key={section.title} className="sidebar-section">
            {!collapsed && (
              <div className="sidebar-section-title">{section.title}</div>
            )}
            {section.items.map((item) => {
              const isActive = pathname === item.href || pathname.startsWith(item.href + "/");
              return (
                <div key={item.href}>
                  <Link
                    href={item.href}
                    className={`sidebar-link ${isActive ? "sidebar-link-active" : ""}`}
                    title={collapsed ? item.label : undefined}
                  >
                    <span className="sidebar-link-icon">{item.icon}</span>
                    {!collapsed && <span className="sidebar-link-label">{item.label}</span>}
                  </Link>
                  {!collapsed && item.children && item.children.length > 0 && (
                    <div className="sidebar-children">
                      {item.children.map((child) => {
                        const isChildActive = pathname === child.href;
                        return (
                          <Link
                            key={child.href}
                            href={child.href}
                            className={`sidebar-link sidebar-child-link ${isChildActive ? "sidebar-link-active" : ""}`}
                          >
                            {child.label}
                          </Link>
                        );
                      })}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        ))}
      </nav>

      <div className="sidebar-footer">
        <div className="sidebar-user">
          <div className="sidebar-user-avatar">
            {user?.username?.charAt(0).toUpperCase() || "U"}
          </div>
          {!collapsed && (
            <div className="sidebar-user-info">
              <span className="sidebar-user-name">{user?.username}</span>
              <span className="sidebar-user-role">Administrator</span>
            </div>
          )}
        </div>
        <button
          onClick={handleLogout}
          className="sidebar-logout"
          title="Logout"
        >
          ⏻
        </button>
      </div>
    </aside>
  );
}
