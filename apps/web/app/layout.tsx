import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Skill Hub EE",
  description: "Enterprise operating system for governed intelligence execution.",
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
              <img src="/logo.svg" alt="Skill Hub EE" className="nav-logo" />
              <span>Skill Hub</span>
              <span className="nav-edition">EE</span>
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

