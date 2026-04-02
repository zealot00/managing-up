"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "../context/AuthContext";
import { useRouter } from "next/navigation";

const navLinks = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/seh", label: "SEH" },
  { href: "/skills", label: "Skills" },
  { href: "/executions", label: "Executions" },
  { href: "/approvals", label: "Approvals" },
  { href: "/tasks", label: "Tasks" },
  { href: "/evaluations", label: "Evaluations" },
  { href: "/experiments", label: "Experiments" },
  { href: "/replays", label: "Replays" },
];

export default function NavBar() {
  const pathname = usePathname();
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

  return (
    <nav className="nav-bar">
      <div className="nav-inner">
        <a href="/" className="nav-brand">
          <img src="/logo.svg" alt="managing-up logo" className="nav-logo" />
          <span>managing-up</span>
          <span className="nav-edition">EE</span>
        </a>
        <div className="nav-links">
          {navLinks.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className={`nav-link${pathname === link.href ? " active" : ""}`}
            >
              {link.label}
            </Link>
          ))}
        </div>
        <div className="nav-user">
          <span className="nav-username">{user?.username}</span>
          <button onClick={handleLogout} className="nav-logout">
            Logout
          </button>
        </div>
      </div>
    </nav>
  );
}
