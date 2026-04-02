import { Suspense } from "react";
import Link from "next/link";
import { getSkills, getSkillVersions } from "../lib/api";
import CreateSkillForm from "../components/CreateSkillForm";
import CreateSkillVersionForm from "../components/CreateSkillVersionForm";

function SkeletonSkillsPage() {
  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Registry</p>
        <h1>Skill inventory and version posture.</h1>
        <p className="lede">
          Track registry ownership, risk classification, and publish state across the current automation surface.
        </p>
      </header>

      <div className="form-panel">
        <div className="loading-pulse loading-pulse-short" style={{ marginBottom: 16 }} />
        <div className="form-fields">
          <div className="loading-pulse loading-pulse-medium" />
          <div className="loading-pulse loading-pulse-medium" />
          <div className="loading-pulse loading-pulse-short" />
        </div>
      </div>

      <div className="grid">
        <div className="panel">
          <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 16 }} />
          <div className="skeleton-grid">
            {[1, 2, 3].map((i) => <div key={i} className="skeleton-card" />)}
          </div>
        </div>
        <div className="panel">
          <div className="loading-pulse loading-pulse-medium" style={{ marginBottom: 16 }} />
          <div className="skeleton-grid">
            {[1, 2, 3].map((i) => <div key={i} className="skeleton-card" />)}
          </div>
        </div>
      </div>
    </main>
  );
}

async function SkillsContent() {
  const [skills, versions] = await Promise.all([getSkills(), getSkillVersions()]);

  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">Registry</p>
        <h1>Skill inventory and version posture.</h1>
        <p className="lede">
          Track registry ownership, risk classification, and publish state across the current automation surface.
        </p>
      </header>

      <CreateSkillForm />

      <CreateSkillVersionForm skills={skills.items} />

      <div className="panel-grid panel-grid-wide">
        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">Registry</p>
            <h2 className="panel-title">Skill entries</h2>
          </div>
          <div className="list">
            {skills.items.length === 0 ? (
              <p className="empty-note">No skills in registry</p>
            ) : (
              skills.items.map((skill) => (
                <Link
                  href={`/skills/${skill.id}`}
                  key={skill.id}
                  className="list-card"
                  style={{ textDecoration: "none" }}
                >
                  <div className="list-card-main">
                    <h3 className="list-card-title">{skill.name}</h3>
                    <p className="list-card-meta">
                      {skill.owner_team} · {skill.risk_level} risk · {skill.current_version || "no published version"}
                    </p>
                  </div>
                  <span className={`badge badge-${skill.status}`}>{skill.status}</span>
                </Link>
              ))
            )}
          </div>
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">Versions</p>
            <h2 className="panel-title">Release history</h2>
          </div>
          <div className="list">
            {versions.items.length === 0 ? (
              <p className="empty-note">No versions yet</p>
            ) : (
              versions.items.map((version) => (
                <article className="list-card" key={version.id}>
                  <div className="list-card-main">
                    <h3 className="list-card-title">
                      {version.skill_id} · {version.version}
                    </h3>
                    <p className="list-card-meta">{version.change_summary}</p>
                  </div>
                  <span className={`badge badge-${version.status}`}>{version.status}</span>
                </article>
              ))
            )}
          </div>
        </article>
      </div>
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
