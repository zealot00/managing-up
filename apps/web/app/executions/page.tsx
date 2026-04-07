import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getExecutions, getSkills } from "../lib/api";
import ExecutionsPageClient from "../components/ExecutionsPageClient";

function SkeletonExecutionsPage() {
  return (
    <main className="shell">
      <header className="hero-page hero-compact" style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
        <div>
          <p className="eyebrow">Execution Timeline</p>
          <h1>Operational runs across governed skills.</h1>
          <p className="lede">
            Monitor live and completed runs, current step progression, and operator-triggered events.
          </p>
        </div>
        <div className="loading-pulse" style={{ width: 160, height: 40, borderRadius: "var(--radius-sm)" }} />
      </header>

      <div style={{ display: "flex", gap: "var(--space-4)", marginBottom: "var(--space-6)" }}>
        <div className="loading-pulse" style={{ width: 200, height: 40 }} />
        <div style={{ display: "flex", gap: "var(--space-2)" }}>
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="loading-pulse" style={{ width: 70, height: 32, borderRadius: "var(--radius-sm)" }} />
          ))}
        </div>
      </div>

      <div className="panel">
        <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 16 }} />
        <div className="skeleton-grid">
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="skeleton-card" />
          ))}
        </div>
      </div>
    </main>
  );
}

async function ExecutionsContent() {
  const [executions, skills] = await Promise.all([getExecutions(), getSkills()]);

  return (
    <main className="shell">
      <ExecutionsPageClient executions={executions} skills={skills.items} />
    </main>
  );
}

export default function ExecutionsPage() {
  return (
    <Suspense fallback={<SkeletonExecutionsPage />}>
      <ExecutionsContent />
    </Suspense>
  );
}
