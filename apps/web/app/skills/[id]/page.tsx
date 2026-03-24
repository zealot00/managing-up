import { notFound } from "next/navigation";
import Link from "next/link";
import { getSkill, getSkillVersions } from "../../lib/api";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function SkillDetailPage({ params }: Props) {
  const { id } = await params;

  let skill;
  try {
    skill = await getSkill(id);
  } catch {
    notFound();
  }

  const versionsData = await getSkillVersions();
  const versions = versionsData.items.filter((v) => v.skill_id === id);

  return (
    <main className="shell">
      <section className="toprail">
        <Link href="/skills" className="toprail-link">
          ← Back to Registry
        </Link>
      </section>

      <section className="hero hero-compact">
        <p className="eyebrow">Registry</p>
        <h1>{skill.name}</h1>
        <p className="lede">
          {skill.owner_team} · {skill.risk_level} risk ·{" "}
          <span className={`badge badge-${skill.status}`}>{skill.status}</span>
        </p>
      </section>

      <section className="panel-grid panel-grid-wide">
        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">Skill</p>
            <h2>Metadata</h2>
          </div>
          <div className="detail-grid">
            <div className="detail-row">
              <span className="detail-label">ID</span>
              <span className="detail-value">{skill.id}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Name</span>
              <span className="detail-value">{skill.name}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Owner team</span>
              <span className="detail-value">{skill.owner_team}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Risk level</span>
              <span className="detail-value">{skill.risk_level}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Status</span>
              <span className="detail-value">{skill.status}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Current version</span>
              <span className="detail-value">{skill.current_version || "—"}</span>
            </div>
          </div>
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">Versions</p>
            <h2>Release history ({versions.length})</h2>
          </div>
          <div className="list">
            {versions.length === 0 ? (
              <p className="empty-note">No versions published yet.</p>
            ) : (
              versions.map((version) => (
                <article className="list-card" key={version.id}>
                  <div>
                    <h3>
                      {version.version} · {version.status}
                    </h3>
                    <p>{version.change_summary}</p>
                    <p className="meta">
                      {version.approval_required ? "Approval required" : "No approval"} ·{" "}
                      {new Date(version.created_at).toLocaleDateString()}
                    </p>
                  </div>
                  <span className={`badge badge-${version.status}`}>{version.status}</span>
                </article>
              ))
            )}
          </div>
        </article>
      </section>
    </main>
  );
}
