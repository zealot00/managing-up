"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "../context/AuthContext";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";

const navLinks = [
  { href: "/dashboard", labelKey: "dashboard" },
  { href: "/seh", labelKey: "sehModule" },
  { href: "/skills", labelKey: "skills" },
  { href: "/executions", labelKey: "executions" },
  { href: "/approvals", labelKey: "approvals" },
  { href: "/tasks", labelKey: "tasks" },
  { href: "/gateway", labelKey: "gateway" },
  { href: "/evaluations", labelKey: "evaluations" },
  { href: "/experiments", labelKey: "experiments" },
  { href: "/replays", labelKey: "replays" },
];

export default function NavBar() {
  const t = useTranslations("nav");
  const tc = useTranslations("common");
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
            <span>{tc("appName")}</span>
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
            <span>{tc("appName")}</span>
          </a>
          <div className="nav-links" style={{ marginLeft: "auto" }}>
            <Link href="/login" className="nav-link">
              {tc("login")}
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
          <span>{tc("appName")}</span>
          <span className="nav-edition">EE</span>
        </a>
        <div className="nav-links">
          {navLinks.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className={`nav-link${pathname === link.href ? " active" : ""}`}
            >
              {t(link.labelKey)}
            </Link>
          ))}
        </div>
        <div className="nav-user">
          <span className="nav-username">{user?.username}</span>
          <button onClick={handleLogout} className="nav-logout">
            {tc("logout")}
          </button>
        </div>
      </div>
    </nav>
  );
}