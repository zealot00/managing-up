"use client";

import { useState, useRef, useEffect, ReactNode } from "react";
import { LogOut, Settings, User, ChevronDown } from "lucide-react";
import { useTranslations } from "next-intl";
import { useRouter } from "next/navigation";

interface UserDropdownProps {
  username: string;
  onLogout: () => void;
  collapsed?: boolean;
}

export default function UserDropdown({ username, onLogout, collapsed = false }: UserDropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const t = useTranslations("userDropdown");
  const router = useRouter();

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape" && isOpen) {
        setIsOpen(false);
      }
    }

    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen]);

  const initials = username?.charAt(0).toUpperCase() || "U";

  return (
    <div ref={dropdownRef} style={{ position: "relative" }}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={`sidebar-user-trigger ${collapsed ? "collapsed" : ""}`}
        style={{
          display: "flex",
          alignItems: "center",
          gap: "var(--space-3)",
          padding: collapsed ? "var(--space-2)" : "var(--space-2) var(--space-3)",
          borderRadius: "var(--radius-sm)",
          background: isOpen ? "var(--sidebar-bg-hover)" : "transparent",
          border: "none",
          cursor: "pointer",
          width: "100%",
          transition: "all var(--transition-fast)",
        }}
      >
        <div
          className="sidebar-user-avatar"
          style={{
            width: collapsed ? 32 : 36,
            height: collapsed ? 32 : 36,
            flexShrink: 0,
          }}
        >
          {initials}
        </div>
        {!collapsed && (
          <>
            <div className="sidebar-user-info" style={{ flex: 1, minWidth: 0 }}>
              <span className="sidebar-user-name" style={{
                whiteSpace: "nowrap",
                overflow: "hidden",
                textOverflow: "ellipsis",
                display: "block",
              }}>
                {username}
              </span>
              <span className="sidebar-user-role">{t("administrator")}</span>
            </div>
            <ChevronDown
              size={16}
              style={{
                color: "var(--sidebar-text)",
                transform: isOpen ? "rotate(180deg)" : "rotate(0)",
                transition: "transform var(--transition-fast)",
                flexShrink: 0,
              }}
            />
          </>
        )}
      </button>

      {isOpen && (
        <div
          className="sidebar-user-dropdown"
          style={{
            position: "absolute",
            bottom: 0,
            left: "100%",
            marginLeft: 8,
            minWidth: 200,
            background: "var(--sidebar-bg-hover)",
            border: "1px solid var(--sidebar-border)",
            borderRadius: "var(--radius-md)",
            boxShadow: "var(--shadow-lg)",
            overflow: "hidden",
            zIndex: 300,
            animation: "slideUp 0.15s ease-out",
          }}
        >
          <div style={{
            padding: "var(--space-3) var(--space-4)",
            borderBottom: "1px solid var(--sidebar-border)",
          }}>
            <p style={{ fontSize: "var(--text-sm)", fontWeight: 600, color: "var(--sidebar-text-active)" }}>
              {username}
            </p>
            <p style={{ fontSize: "var(--text-xs)", color: "var(--sidebar-text)" }}>
              {t("administrator")}
            </p>
          </div>

          <div style={{ padding: "var(--space-2)" }}>
            <button
              onClick={() => {
                setIsOpen(false);
                router.push("/profile");
              }}
              style={{
                display: "flex",
                alignItems: "center",
                gap: "var(--space-3)",
                width: "100%",
                padding: "var(--space-2) var(--space-3)",
                borderRadius: "var(--radius-sm)",
                background: "transparent",
                border: "none",
                cursor: "pointer",
                fontSize: "var(--text-sm)",
                color: "var(--sidebar-text)",
                transition: "background var(--transition-fast)",
              }}
              onMouseEnter={(e) => e.currentTarget.style.background = "var(--sidebar-bg-active)"}
              onMouseLeave={(e) => e.currentTarget.style.background = "transparent"}
            >
              <User size={16} style={{ color: "var(--sidebar-text)" }} />
              <span>{t("profileSettings")}</span>
            </button>

            <button
              onClick={() => {
                setIsOpen(false);
                router.push("/preferences");
              }}
              style={{
                display: "flex",
                alignItems: "center",
                gap: "var(--space-3)",
                width: "100%",
                padding: "var(--space-2) var(--space-3)",
                borderRadius: "var(--radius-sm)",
                background: "transparent",
                border: "none",
                cursor: "pointer",
                fontSize: "var(--text-sm)",
                color: "var(--sidebar-text)",
                transition: "background var(--transition-fast)",
              }}
              onMouseEnter={(e) => e.currentTarget.style.background = "var(--sidebar-bg-active)"}
              onMouseLeave={(e) => e.currentTarget.style.background = "transparent"}
            >
              <Settings size={16} style={{ color: "var(--sidebar-text)" }} />
              <span>{t("preferences")}</span>
            </button>
          </div>

          <div style={{
            borderTop: "1px solid var(--sidebar-border)",
            padding: "var(--space-2)",
          }}>
            <button
              onClick={() => {
                setIsOpen(false);
                onLogout();
              }}
              style={{
                display: "flex",
                alignItems: "center",
                gap: "var(--space-3)",
                width: "100%",
                padding: "var(--space-2) var(--space-3)",
                borderRadius: "var(--radius-sm)",
                background: "transparent",
                border: "none",
                cursor: "pointer",
                fontSize: "var(--text-sm)",
                color: "var(--danger)",
                transition: "background var(--transition-fast)",
              }}
              onMouseEnter={(e) => e.currentTarget.style.background = "rgba(185, 28, 28, 0.12)"}
              onMouseLeave={(e) => e.currentTarget.style.background = "transparent"}
            >
              <LogOut size={16} />
              <span>{t("logout")}</span>
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
