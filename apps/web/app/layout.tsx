import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "managing-up — 向上管理",
  description: "Enterprise AI quality infrastructure — benchmark, regression, and harness testing for AI agents.",
};

const navLinks = [
  { href: "/skills", label: "Skills" },
  { href: "/executions", label: "Executions" },
  { href: "/approvals", label: "Approvals" },
  { href: "/tasks", label: "Tasks" },
  { href: "/evaluations", label: "Evaluations" },
  { href: "/experiments", label: "Experiments" },
  { href: "/replays", label: "Replays" },
];

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <nav className="nav-bar">
          <div className="nav-inner">
            <a href="/" className="nav-brand">
              <img src="/logo.svg" alt="managing-up" className="nav-logo" />
              <span>managing-up</span>
              <span className="nav-edition">向上管理</span>
            </a>
            <div className="nav-links">
              {navLinks.map((link) => (
                <a key={link.href} href={link.href} className="nav-link">
                  {link.label}
                </a>
              ))}
            </div>
          </div>
        </nav>
        {children}
      </body>
    </html>
  );
}

