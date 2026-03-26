import type { Metadata } from "next";
import "./globals.css";
import NavBar from "../components/NavBar";
import { AuthProvider } from "../context/AuthContext";

export const metadata: Metadata = {
  title: "managing-up — 向上管理",
  description: "Enterprise AI quality infrastructure — benchmark, regression, and harness testing for AI agents.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <AuthProvider>
          <NavBar />
          {children}
        </AuthProvider>
      </body>
    </html>
  );
}
