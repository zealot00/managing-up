import { Suspense } from "react";
import { getSkills, getSkillVersions } from "../lib/api";
import CreateSkillForm from "../components/CreateSkillForm";
import { SkeletonPanel } from "../components/SkeletonPanel";

function SkeletonSkillsPage() {
  return (
    <main className="shell">
      <section className="hero hero-compact">
        <p className="eyebrow">Registry</p>
        <h1>Skill inventory and version posture.</h1>
        <p className="lede">
          Track registry ownership, risk classification, and publish state across the current
          automation surface.
        </p>
      </section>

      <div className="form-panel">
        <div className="loading-pulse loading-pulse-short" style={{ marginBottom: 12 }} />
        <div className="form-fields">
          <div className="loading-pulse loading-pulse-medium" />
          <div className="loading-pulse loading-pulse-medium" />
          <div className="loading-pulse loading-pulse-short" />
        </div>
        <div className="loading-pulse" style={{ width: 140, height: 44, borderRadius: 999 }} />
      </div>

      <section className="panel-grid panel-grid-wide">
        <SkeletonPanel height={320} />
        <SkeletonPanel height={320} />
      </section>
    </main>
  );
}

async function SkillsContent() {
  const [skills, versions] = await Promise.all([getSkills(), getSkillVersions()]);

  return (
    <main className="shell">
      <section className="hero hero-compact">
        <p className="eyebrow">Registry</p>
        <h1>Skill inventory and version posture.</h1>
        <p className="lede">
          Track registry ownership, risk classification, and publish state across the current
          automation surface.
        </p>
      </section>

      <CreateSkillForm />

      <section className="panel-grid panel-grid-wide">
        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">Skills</p>
            <h2>Registry entries</h2>
          </div>
          <div className="list">
            {skills.items.map((skill) => (
              <article className="list-card" key={skill.id}>
                <div>
                  <h3>{skill.name}</h3>
                  <p>
                    {skill.owner_team} · {skill.risk_level} risk · {skill.current_version || "no published version"}
                  </p>
                </div>
                <span className={`badge badge-${skill.status}`}>{skill.status}</span>
              </article>
            ))}
          </div>
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">Versions</p>
            <h2>Release history</h2>
          </div>
          <div className="list">
            {versions.items.map((version) => (
              <article className="list-card" key={version.id}>
                <div>
                  <h3>
                    {version.skill_id} · {version.version}
                  </h3>
                  <p>{version.change_summary}</p>
                </div>
                <span className={`badge badge-${version.status}`}>{version.status}</span>
              </article>
            ))}
          </div>
        </article>
      </section>
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
