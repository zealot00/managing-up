import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { getSkills } from "../lib/api";
import SkillsPageClient from "../components/SkillsPageClient";

function SkeletonSkillsPage() {
  return (
    <main className="shell">
      <header className="hero-page hero-compact" style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
        <div>
          <p className="eyebrow">Registry</p>
          <h1>Skill inventory and version posture.</h1>
          <p className="lede">
            Track registry ownership, risk classification, and publish state across the current automation surface.
          </p>
        </div>
        <div className="loading-pulse" style={{ width: 140, height: 40, borderRadius: "var(--radius-sm)" }} />
      </header>

      <div className="panel">
        <div className="panel-header">
          <div className="loading-pulse loading-pulse-short" style={{ width: 120, marginBottom: 8 }} />
          <div className="loading-pulse loading-pulse-medium" style={{ width: 200 }} />
        </div>
        <div className="skeleton-grid">
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="skeleton-card" />
          ))}
        </div>
      </div>
    </main>
  );
}

async function SkillsContent() {
  const skills = await getSkills();

  return (
    <main className="shell">
      <SkillsPageClient skills={skills} />
    </main>
  );
}

export default function SkillsPage() {
  return (
    <Suspense fallback={<SkeletonSkillsPage />}>
      <SkillsContent />
    </Suspense>
  );
}
