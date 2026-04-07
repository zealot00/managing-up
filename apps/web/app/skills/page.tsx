import { Suspense } from "react";
import Link from "next/link";
import { getTranslations } from "next-intl/server";
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
  const t = await getTranslations("skills");
  const [skills, versions] = await Promise.all([getSkills(), getSkillVersions()]);

  return (
    <main className="shell">
      <header className="hero-page hero-compact">
        <p className="eyebrow">{t("eyebrow")}</p>
        <h1>{t("title")}</h1>
        <p className="lede">{t("lede")}</p>
      </header>

      <CreateSkillForm />

      <CreateSkillVersionForm skills={skills.items} />

      <div className="panel-grid panel-grid-wide">
        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">{t("eyebrow")}</p>
            <h2 className="panel-title">{t("title")}</h2>
          </div>
          <div className="table-wrapper">
            {skills.items.length === 0 ? (
              <p className="empty-note">{t("noSkills")}</p>
            ) : (
              <table className="table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Owner</th>
                    <th>Risk Level</th>
                    <th>Version</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {skills.items.map((skill) => (
                    <tr key={skill.id} style={{ cursor: "pointer" }}>
                      <td>
                        <Link href={`/skills/${skill.id}`} style={{ textDecoration: "none" }}>
                          {skill.name}
                        </Link>
                      </td>
                      <td>{skill.owner_team}</td>
                      <td>{skill.risk_level}</td>
                      <td>{skill.current_version || t("noVersions")}</td>
                      <td>
                        <span className={`badge badge-${skill.status}`}>{skill.status}</span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">{t("versions")}</p>
            <h2 className="panel-title">{t("versions")}</h2>
          </div>
          <div className="table-wrapper">
            {versions.items.length === 0 ? (
              <p className="empty-note">{t("noVersions")}</p>
            ) : (
              <table className="table">
                <thead>
                  <tr>
                    <th>Skill ID</th>
                    <th>Version</th>
                    <th>Change Summary</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {versions.items.map((version) => (
                    <tr key={version.id}>
                      <td>{version.skill_id}</td>
                      <td>{version.version}</td>
                      <td>{version.change_summary}</td>
                      <td>
                        <span className={`badge badge-${version.status}`}>{version.status}</span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
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
