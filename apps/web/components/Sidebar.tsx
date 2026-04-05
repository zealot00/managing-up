"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "../context/AuthContext";
import { useRouter } from "next/navigation";
import { useState, useEffect } from "react";
import { useTranslations } from "next-intl";
import LanguageSwitcher from "./LanguageSwitcher";
import { useMobileSidebar } from "./MobileSidebarProvider";
import {
  LayoutDashboard,
  Package,
  Play,
  ListChecks,
  Target,
  FlaskConical,
  ClipboardCheck,
  RotateCcw,
  Network,
  Shield,
  LogOut,
  PanelLeftClose,
  PanelLeftOpen,
} from "lucide-react";

interface NavItem {
  href: string;
  labelKey: string;
  icon: React.ReactNode;
  children?: { href: string; labelKey: string }[];
}

export default function Sidebar() {
  const t = useTranslations("nav");
  const { isOpen, close } = useMobileSidebar();

  const navSections: { titleKey: string; items: NavItem[] }[] = [
    {
      titleKey: "dashboard",
      items: [
        { href: "/dashboard", labelKey: "dashboard", icon: <LayoutDashboard size={18} aria-hidden="true" /> },
      ],
    },
    {
      titleKey: "aiQuality",
      items: [
        { href: "/skills", labelKey: "skills", icon: <Package size={18} aria-hidden="true" /> },
        { href: "/executions", labelKey: "executions", icon: <Play size={18} aria-hidden="true" /> },
        { href: "/tasks", labelKey: "tasks", icon: <ListChecks size={18} aria-hidden="true" />, children: [{ href: "/tasks/from-trace", labelKey: "taskBuilder" }] },
        { href: "/evaluations", labelKey: "evaluations", icon: <Target size={18} aria-hidden="true" /> },
        { href: "/experiments", labelKey: "experiments", icon: <FlaskConical size={18} aria-hidden="true" /> },
      ],
    },
    {
      titleKey: "operations",
      items: [
        { href: "/approvals", labelKey: "approvals", icon: <ClipboardCheck size={18} aria-hidden="true" /> },
        { href: "/replays", labelKey: "replays", icon: <RotateCcw size={18} aria-hidden="true" /> },
        { href: "/gateway", labelKey: "gateway", icon: <Network size={18} aria-hidden="true" /> },
        { href: "/mcp", labelKey: "mcp", icon: <Shield size={18} aria-hidden="true" /> },
      ],
    },
    {
      titleKey: "system",
      items: [
        {
          href: "/seh",
          labelKey: "sehModule",
          icon: <Shield size={18} aria-hidden="true" />,
          children: [
            { href: "/seh/datasets", labelKey: "sehDatasets" },
            { href: "/seh/runs", labelKey: "sehRuns" },
            { href: "/seh/policies", labelKey: "sehPolicies" },
          ],
        },
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

  function handleLinkClick() {
    if (isOpen) {
      close();
    }
  }

  return (
    <>
      {isOpen && <div className="sidebar-backdrop" onClick={close} />}
      <aside className={`sidebar ${collapsed ? "sidebar-collapsed" : ""} ${isOpen ? "sidebar-mobile-open" : ""}`}>
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
            {collapsed ? <PanelLeftOpen size={18} aria-hidden="true" /> : <PanelLeftClose size={18} aria-hidden="true" />}
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
                    onClick={handleLinkClick}
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
          aria-label="Logout"
        >
          <LogOut size={18} aria-hidden="true" />
        </button>
      </div>
    </aside>
    </>
  );
}
