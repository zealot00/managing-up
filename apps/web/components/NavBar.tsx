"use client";

import Link from "next/link";
import { useAuth } from "../context/AuthContext";
import { useRouter } from "next/navigation";

const navLinks = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/skills", label: "Skills" },
  { href: "/executions", label: "Executions" },
  { href: "/approvals", label: "Approvals" },
  { href: "/tasks", label: "Tasks" },
  { href: "/evaluations", label: "Evaluations" },
  { href: "/experiments", label: "Experiments" },
  { href: "/replays", label: "Replays" },
];

export default function NavBar() {
  const { user, isAuthenticated, isLoading, logout } = useAuth();
  const router = useRouter();

  async function handleLogout() {
    await logout();
    router.push("/login");
  }

  if (isLoading) {
    return (
      <nav className="nav-bar">
        <div className="nav-inner">
          <a href="/" className="nav-brand">
            <img src="/logo.svg" alt="managing-up logo" className="nav-logo" />
            <span>managing-up</span>
          </a>
        </div>
      </nav>
    );
  }

  if (!isAuthenticated) {
    // Minimal nav for logged-out users
    return (
      <nav className="nav-bar">
        <div className="nav-inner">
          <a href="/" className="nav-brand">
            <img src="/logo.svg" alt="managing-up logo" className="nav-logo" />
            <span>managing-up</span>
          </a>
          <div className="nav-links" style={{ marginLeft: "auto" }}>
            <Link href="/login" className="nav-link">
              Login
            </Link>
          </div>
        </div>
      </nav>
    );
  }

  // Full nav for logged-in users
  return (
    <nav className="nav-bar">
      <div className="nav-inner">
        <a href="/" className="nav-brand">
          <img src="/logo.svg" alt="managing-up logo" className="nav-logo" />
          <span>managing-up</span>
          <span className="nav-edition">向上管理</span>
        </a>
        <div className="nav-links">
          {navLinks.map((link) => (
            <Link key={link.href} href={link.href} className="nav-link">
              {link.label}
            </Link>
          ))}
        </div>
        <div style={{ marginLeft: "auto", display: "flex", alignItems: "center", gap: "12px" }}>
          <span style={{ fontSize: "0.85rem", color: "var(--muted)" }}>{user?.username}</span>
          <button onClick={handleLogout} className="nav-link" style={{ background: "none", border: "none", cursor: "pointer" }}>
            Logout
          </button>
        </div>
      </div>
    </nav>
  );
}
