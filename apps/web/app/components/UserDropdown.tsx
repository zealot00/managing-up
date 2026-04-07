"use client";

import { useState, useRef, useEffect, ReactNode } from "react";
import { LogOut, Settings, User, ChevronDown } from "lucide-react";

interface UserDropdownProps {
  username: string;
  onLogout: () => void;
  collapsed?: boolean;
}

export default function UserDropdown({ username, onLogout, collapsed = false }: UserDropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const initials = username?.charAt(0).toUpperCase() || "U";

  return (
    <div ref={dropdownRef} className="relative">
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
              <span className="sidebar-user-role">Administrator</span>
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
            bottom: "100%",
            left: collapsed ? "50%" : 0,
            right: collapsed ? "auto" : 0,
            transform: collapsed ? "translateX(-50%)" : "translateX(0)",
            marginBottom: 8,
            minWidth: 200,
            background: "var(--surface-raised)",
            border: "1px solid var(--line)",
            borderRadius: "var(--radius-md)",
            boxShadow: "var(--shadow-lg)",
            overflow: "hidden",
            zIndex: 300,
            animation: "slideUp 0.15s ease-out",
          }}
        >
          <div style={{
            padding: "var(--space-3) var(--space-4)",
            borderBottom: "1px solid var(--line)",
          }}>
            <p style={{ fontSize: "var(--text-sm)", fontWeight: 600, color: "var(--ink-strong)" }}>
              {username}
            </p>
            <p style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>
              Administrator
            </p>
          </div>

          <div style={{ padding: "var(--space-2)" }}>
            <button
              onClick={() => {
                setIsOpen(false);
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
                color: "var(--ink)",
                transition: "background var(--transition-fast)",
              }}
              onMouseEnter={(e) => e.currentTarget.style.background = "var(--bg)"}
              onMouseLeave={(e) => e.currentTarget.style.background = "transparent"}
            >
              <User size={16} style={{ color: "var(--muted)" }} />
              <span>Profile Settings</span>
            </button>

            <button
              onClick={() => {
                setIsOpen(false);
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
                color: "var(--ink)",
                transition: "background var(--transition-fast)",
              }}
              onMouseEnter={(e) => e.currentTarget.style.background = "var(--bg)"}
              onMouseLeave={(e) => e.currentTarget.style.background = "transparent"}
            >
              <Settings size={16} style={{ color: "var(--muted)" }} />
              <span>Preferences</span>
            </button>
          </div>

          <div style={{
            borderTop: "1px solid var(--line)",
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
              onMouseEnter={(e) => e.currentTarget.style.background = "rgba(153, 27, 27, 0.08)"}
              onMouseLeave={(e) => e.currentTarget.style.background = "transparent"}
            >
              <LogOut size={16} />
              <span>Log out</span>
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
