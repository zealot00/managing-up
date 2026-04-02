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
      <div className="landing-page">
        <section className="landing-hero">
          <div className="loading-pulse loading-pulse-short" style={{ width: 200, marginBottom: 16 }} />
          <div className="loading-pulse loading-pulse-medium" style={{ width: 400, marginBottom: 16 }} />
          <div className="loading-pulse loading-pulse-long" style={{ width: 600 }} />
        </section>
      </div>
    );
  }

  if (isAuthenticated) {
    return null;
  }

  return (
    <div className="landing-page">
      <section className="landing-hero">
        <p className="landing-eyebrow">Enterprise AI Quality Infrastructure</p>
        <h1 className="landing-title">Managing Up</h1>
        <p className="landing-subtitle">
          Benchmark testing, regression detection, and harness evaluation for AI agents.
          Full audit trails, human approvals, and quantitative reporting.
        </p>
        <div className="landing-actions">
          <a href="/login" className="landing-btn landing-btn-primary">Get Started</a>
          <a href="https://github.com/your-org/managing-up" className="landing-btn landing-btn-secondary" target="_blank" rel="noopener noreferrer">View on GitHub</a>
        </div>
      </section>

      <section className="landing-features">
        <div className="landing-features-grid">
          <article className="landing-feature">
            <div className="landing-feature-icon">◈</div>
            <h3 className="landing-feature-title">Skill Registry</h3>
            <p className="landing-feature-desc">Version-controlled SOPs as executable skills with approval workflows.</p>
          </article>
          <article className="landing-feature">
            <div className="landing-feature-icon">▸</div>
            <h3 className="landing-feature-title">Execution Engine</h3>
            <p className="landing-feature-desc">State machine with checkpoints, human approvals, and full trace capture.</p>
          </article>
          <article className="landing-feature">
            <div className="landing-feature-icon">◎</div>
            <h3 className="landing-feature-title">Evaluation Pipeline</h3>
            <p className="landing-feature-desc">Multiple evaluator types with regression detection and A/B comparison.</p>
          </article>
          <article className="landing-feature">
            <div className="landing-feature-icon">↻</div>
            <h3 className="landing-feature-title">Trace Replay</h3>
            <p className="landing-feature-desc">Deterministic reproduction of AI agent behaviors for debugging.</p>
          </article>
        </div>
      </section>
    </div>
  );
}
