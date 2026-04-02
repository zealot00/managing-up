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
          <div className="loading-pulse loading-pulse-short" style={{ width: 200, marginBottom: 16 }} />
          <div className="loading-pulse loading-pulse-medium" style={{ width: 400, marginBottom: 16 }} />
          <div className="loading-pulse loading-pulse-long" style={{ width: 600 }} />
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

      <div className="grid" style={{ marginTop: "var(--space-6)" }}>
        <article className="card card-interactive" onClick={() => window.open("https://github.com/your-org/managing-up", "_blank")}>
          <h2 style={{ fontSize: "var(--text-xl)", marginBottom: "var(--space-3)" }}>Get Started</h2>
          <p style={{ color: "var(--muted)" }}>Clone the repository and start building your AI quality infrastructure today.</p>
          <span style={{ color: "var(--primary)", fontWeight: 600, marginTop: "var(--space-4)", display: "inline-block" }}>GitHub →</span>
        </article>
        
        <article className="card card-interactive" onClick={() => window.open("https://docs.example.com", "_blank")}>
          <h2 style={{ fontSize: "var(--text-xl)", marginBottom: "var(--space-3)" }}>Documentation</h2>
          <p style={{ color: "var(--muted)" }}>Comprehensive guides, API references, and integration examples.</p>
          <span style={{ color: "var(--primary)", fontWeight: 600, marginTop: "var(--space-4)", display: "inline-block" }}>Read Docs →</span>
        </article>
      </div>

      <section style={{ marginTop: "var(--space-8)" }}>
        <div className="panel-header" style={{ marginBottom: "var(--space-5)" }}>
          <p className="section-kicker">Ecosystem</p>
          <h2 className="panel-title">Platform Modules</h2>
        </div>
        <div className="grid grid-3">
          <article className="card">
            <h3 style={{ fontSize: "var(--text-lg)", marginBottom: "var(--space-3)", color: "var(--ink-strong)" }}>SEH - Skill Eval Harness</h3>
            <p style={{ color: "var(--muted)", fontSize: "var(--text-sm)" }}>
              Standardized evaluation framework for testing AI agent skills against defined test cases.
            </p>
          </article>
          <article className="card">
            <h3 style={{ fontSize: "var(--text-lg)", marginBottom: "var(--space-3)", color: "var(--ink-strong)" }}>Skill Registry</h3>
            <p style={{ color: "var(--muted)", fontSize: "var(--text-sm)" }}>
              Version-controlled SOPs as executable skills with approval workflows.
            </p>
          </article>
          <article className="card">
            <h3 style={{ fontSize: "var(--text-lg)", marginBottom: "var(--space-3)", color: "var(--ink-strong)" }}>Trace & Replay</h3>
            <p style={{ color: "var(--muted)", fontSize: "var(--text-sm)" }}>
              Full execution traces and deterministic replay for debugging AI behaviors.
            </p>
          </article>
          <article className="card">
            <h3 style={{ fontSize: "var(--text-lg)", marginBottom: "var(--space-3)", color: "var(--ink-strong)" }}>Benchmark Engine</h3>
            <p style={{ color: "var(--muted)", fontSize: "var(--text-sm)" }}>
              Quantitative performance metrics across different models and configurations.
            </p>
          </article>
          <article className="card">
            <h3 style={{ fontSize: "var(--text-lg)", marginBottom: "var(--space-3)", color: "var(--ink-strong)" }}>Approval Gates</h3>
            <p style={{ color: "var(--muted)", fontSize: "var(--text-sm)" }}>
              Human-in-the-loop checkpoints for high-risk operations and policy enforcement.
            </p>
          </article>
          <article className="card">
            <h3 style={{ fontSize: "var(--text-lg)", marginBottom: "var(--space-3)", color: "var(--ink-strong)" }}>Experiment Tracking</h3>
            <p style={{ color: "var(--muted)", fontSize: "var(--text-sm)" }}>
              A/B comparison runs for skills, agents, and model configurations.
            </p>
          </article>
        </div>
      </section>
    </main>
  );
}
