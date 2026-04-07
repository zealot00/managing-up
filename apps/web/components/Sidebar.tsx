"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "../context/AuthContext";
import { useRouter } from "next/navigation";
import { useState, useRef } from "react";
import { useTranslations } from "next-intl";
import { useMobileSidebar } from "./MobileSidebarProvider";
import UserDropdown from "../app/components/UserDropdown";
import Tooltip from "../app/components/Tooltip";
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
  PanelLeftClose,
  PanelLeftOpen,
  ChevronRight,
  ChevronDown,
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
        { href: "/gateway", labelKey: "gateway", icon: <Network size={18} aria-hidden="true" />, children: [{ href: "/gateway/providers", labelKey: "providers" }] },
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
  const { user, isAuthenticated, logout } = useAuth();
  const router = useRouter();
  const [collapsed, setCollapsed] = useState(false);
  const [hoveredItem, setHoveredItem] = useState<string | null>(null);
  const [expandedItems, setExpandedItems] = useState<Set<string>>(new Set());

  async function handleLogout() {
    await logout();
    router.push("/login");
  }

  function toggleExpanded(href: string) {
    setExpandedItems((prev) => {
      const next = new Set(prev);
      if (next.has(href)) {
        next.delete(href);
      } else {
        next.add(href);
      }
      return next;
    });
  }

  function handleLinkClick() {
    if (isOpen) {
      close();
    }
  }

  if (!isAuthenticated) {
    return null;
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
                const hasChildren = item.children && item.children.length > 0;
                const isExpanded = expandedItems.has(item.href);

                if (collapsed) {
                  if (hasChildren) {
                    return (
                      <div
                        key={item.href}
                        className="relative"
                        onMouseEnter={() => setHoveredItem(item.href)}
                        onMouseLeave={() => setHoveredItem(null)}
                      >
                        <Link
                          href={item.href}
                          className={`sidebar-link sidebar-link-collapsed ${isActive ? "sidebar-link-active" : ""}`}
                          title={t(item.labelKey)}
                          onClick={handleLinkClick}
                        >
                          <span className="sidebar-link-icon">{item.icon}</span>
                        </Link>

                        {hoveredItem === item.href && (
                          <div
                            className="sidebar-flyout"
                            style={{
                              position: "absolute",
                              left: "100%",
                              top: 0,
                              minWidth: 180,
                              background: "var(--surface-raised)",
                              border: "1px solid var(--line)",
                              borderRadius: "var(--radius-md)",
                              boxShadow: "var(--shadow-lg)",
                              padding: "var(--space-2)",
                              zIndex: 250,
                              animation: "fadeIn 0.15s ease-out",
                            }}
                          >
                            <div style={{
                              padding: "var(--space-2) var(--space-3)",
                              fontSize: "var(--text-xs)",
                              fontWeight: 600,
                              color: "var(--muted)",
                              textTransform: "uppercase",
                              letterSpacing: "0.05em",
                              borderBottom: "1px solid var(--line)",
                              marginBottom: "var(--space-2)",
                            }}>
                              {t(item.labelKey)}
                            </div>
                            {item.children?.map((child) => {
                              const isChildActive = pathname === child.href;
                              return (
                                <Link
                                  key={child.href}
                                  href={child.href}
                                  className={`sidebar-flyout-item ${isChildActive ? "active" : ""}`}
                                  onClick={handleLinkClick}
                                >
                                  {t(child.labelKey)}
                                </Link>
                              );
                            })}
                          </div>
                        )}
                      </div>
                    );
                  }

                  return (
                    <Tooltip key={item.href} content={t(item.labelKey)} position="right">
                      <Link
                        href={item.href}
                        className={`sidebar-link sidebar-link-collapsed ${isActive ? "sidebar-link-active" : ""}`}
                        title={t(item.labelKey)}
                        onClick={handleLinkClick}
                      >
                        <span className="sidebar-link-icon">{item.icon}</span>
                      </Link>
                    </Tooltip>
                  );
                }

                return (
                  <div key={item.href}>
                    <Link
                      href={item.href}
                      className={`sidebar-link ${isActive ? "sidebar-link-active" : ""}`}
                      onClick={(e) => {
                        if (hasChildren) {
                          e.preventDefault();
                          toggleExpanded(item.href);
                        }
                        handleLinkClick();
                      }}
                    >
                      <span className="sidebar-link-icon">{item.icon}</span>
                      <span className="sidebar-link-label">{t(item.labelKey)}</span>
                      {hasChildren && (
                        <span className="sidebar-link-chevron" style={{ marginLeft: "auto" }}>
                          {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                        </span>
                      )}
                    </Link>

                    {hasChildren && isExpanded && (
                      <div className="sidebar-children" style={{
                        animation: "slideDown 0.2s ease-out",
                      }}>
                        {item.children?.map((child) => {
                          const isChildActive = pathname === child.href;
                          return (
                            <Link
                              key={child.href}
                              href={child.href}
                              className={`sidebar-link sidebar-child-link ${isChildActive ? "sidebar-link-active" : ""}`}
                              onClick={handleLinkClick}
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
          <UserDropdown
            username={user?.username || "User"}
            onLogout={handleLogout}
            collapsed={collapsed}
          />
        </div>
      </aside>
    </>
  );
}
