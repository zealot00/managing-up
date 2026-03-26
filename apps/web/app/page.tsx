"use client";

import { useEffect } from "react";
import { useAuth } from "../context/AuthContext";
import { useRouter } from "next/navigation";

export default function HomePage() {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.replace("/dashboard");
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) {
    return (
      <main className="shell">
        <section className="hero-landing">
          <div className="loading-pulse" style={{ width: 200, height: 20, marginBottom: 16 }} />
          <div className="loading-pulse" style={{ width: 400, height: 40, marginBottom: 16 }} />
          <div className="loading-pulse" style={{ width: 600, height: 20 }} />
        </section>
      </main>
    );
  }

  if (isAuthenticated) {
    return null;
  }

  return (
    <main className="shell">
      <section className="hero-landing">
        <p className="eyebrow">Enterprise AI Quality Infrastructure</p>
        <h1>Convert SOPs into executable skills.</h1>
        <p className="lede">
          Managing-up provides benchmark testing, regression detection, and harness evaluation
          for AI agents. Full audit trails, human approvals, and quantitative reporting.
        </p>
      </section>

      <section className="grid" style={{ marginTop: "24px" }}>
        <article className="card" style={{ cursor: "pointer" }} onClick={() => window.open("https://github.com/your-org/managing-up", "_blank")}>
          <h2>Get Started</h2>
          <p>Clone the repository and start building your AI quality infrastructure today.</p>
          <span style={{ color: "var(--primary)", fontWeight: 600 }}>GitHub →</span>
        </article>
        
        <article className="card" style={{ cursor: "pointer" }} onClick={() => window.open("https://docs.example.com", "_blank")}>
          <h2>Documentation</h2>
          <p>Comprehensive guides, API references, and integration examples.</p>
          <span style={{ color: "var(--primary)", fontWeight: 600 }}>Read Docs →</span>
        </article>
      </section>

      <section style={{ marginTop: "32px" }}>
        <div className="panel-header" style={{ marginBottom: "16px" }}>
          <h2>Ecosystem</h2>
        </div>
        <div className="grid">
          <article className="card">
            <h3>SEH - Skill Eval Harness</h3>
            <p>Standardized evaluation framework for testing AI agent skills against defined test cases.</p>
          </article>
          <article className="card">
            <h3>Skill Registry</h3>
            <p>Version-controlled SOPs as executable skills with approval workflows.</p>
          </article>
          <article className="card">
            <h3>Trace & Replay</h3>
            <p>Full execution traces and deterministic replay for debugging AI behaviors.</p>
          </article>
          <article className="card">
            <h3>Benchmark Engine</h3>
            <p>Quantitative performance metrics across different models and configurations.</p>
          </article>
        </div>
      </section>
    </main>
  );
}
