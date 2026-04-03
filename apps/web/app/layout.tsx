import type { Metadata } from "next";
import "./globals.css";
import Sidebar from "../components/Sidebar";
import AdminHeader from "../components/AdminHeader";
import { AuthProvider } from "../context/AuthContext";
import Providers from "./providers";

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
          <Providers>
            <div className="admin-layout">
              <Sidebar />
              <div className="admin-main">
                <AdminHeader />
                <div className="admin-content">
                  {children}
                </div>
              </div>
            </div>
          </Providers>
        </AuthProvider>
      </body>
    </html>
  );
}
