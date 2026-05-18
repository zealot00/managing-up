import type { Metadata } from "next";
import "./globals.css";
import Sidebar from "../components/Sidebar";
import AdminHeader from "../components/AdminHeader";
import { AuthProvider } from "../context/AuthContext";
import Providers from "./providers";
import { MobileSidebarProvider } from "../components/MobileSidebarProvider";
import { getLocale } from "next-intl/server";

export const metadata: Metadata = {
  title: "managing-up — 向上管理",
  description: "Enterprise AI quality infrastructure — benchmark, regression, and harness testing for AI agents.",
  icons: {
    icon: "/logo.svg",
  },
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const locale = await getLocale();

  return (
    <html lang={locale}>
      <body>
        <a href="#main-content" className="skip-link">Skip to main content</a>
        <AuthProvider>
          <Providers>
            <MobileSidebarProvider>
              <div className="admin-layout">
                <Sidebar />
                <main className="admin-main">
                  <AdminHeader />
                  <div id="main-content" className="admin-content">
                    {children}
                  </div>
                </main>
              </div>
            </MobileSidebarProvider>
          </Providers>
        </AuthProvider>
      </body>
    </html>
  );
}
