"use client";

import LoginForm from "../../components/LoginForm";

export default function LoginPageClient() {
  return (
    <main className="shell" style={{ minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center" }}>
      <div style={{ width: "100%", maxWidth: "440px", padding: "24px" }}>
        <div style={{ textAlign: "center", marginBottom: "32px" }}>
          <h1 style={{ fontSize: "1.5rem", fontWeight: 700, color: "var(--ink-strong)" }}>managing-up</h1>
          <p style={{ color: "var(--muted)", marginTop: "8px" }}>向上管理</p>
        </div>
        <LoginForm />
      </div>
    </main>
  );
}
