"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "../context/AuthContext";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useTranslations } from "next-intl";
import LanguageSwitcher from "./LanguageSwitcher";

interface NavItem {
  href: string;
  labelKey: string;
  icon: string;
  children?: { href: string; labelKey: string }[];
}

export default function Sidebar() {
  const t = useTranslations("nav");

  const navSections: { titleKey: string; items: NavItem[] }[] = [
    {
      titleKey: "dashboard",
      items: [
        { href: "/dashboard", labelKey: "dashboard", icon: "◉" },
      ],
    },
    {
      titleKey: "aiQuality",
      items: [
        { href: "/skills", labelKey: "skills", icon: "◈" },
        { href: "/executions", labelKey: "executions", icon: "▸" },
        { href: "/tasks", labelKey: "tasks", icon: "◉", children: [{ href: "/tasks/from-trace", labelKey: "taskBuilder" }] },
        { href: "/evaluations", labelKey: "evaluations", icon: "◎" },
        { href: "/experiments", labelKey: "experiments", icon: "◇" },
      ],
    },
    {
      titleKey: "operations",
      items: [
        { href: "/approvals", labelKey: "approvals", icon: "◐" },
        { href: "/replays", labelKey: "replays", icon: "↻" },
        { href: "/gateway", labelKey: "gateway", icon: "⬡" },
      ],
    },
    {
      titleKey: "system",
      items: [
        { href: "/seh", labelKey: "sehModule", icon: "▣" },
      ],
    },
  ];

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
          <div key={section.titleKey} className="sidebar-section">
            {!collapsed && (
              <div className="sidebar-section-title">{t(section.titleKey)}</div>
            )}
            {section.items.map((item) => {
              const isActive = pathname === item.href || pathname.startsWith(item.href + "/");
              return (
                <div key={item.href}>
                  <Link
                    href={item.href}
                    className={`sidebar-link ${isActive ? "sidebar-link-active" : ""}`}
                    title={collapsed ? t(item.labelKey) : undefined}
                  >
                    <span className="sidebar-link-icon">{item.icon}</span>
                    {!collapsed && <span className="sidebar-link-label">{t(item.labelKey)}</span>}
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
                            {t(child.labelKey)}
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
        {!collapsed && <LanguageSwitcher />}
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
