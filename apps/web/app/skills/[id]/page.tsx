import { notFound } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { getSkill, getSkillVersions, getSkillSpec } from "../../lib/api";
import Breadcrumb from "../../../components/Breadcrumb";
import { PageHeader } from "../../components/layout/PageHeader";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function SkillDetailPage({ params }: Props) {
  const { id } = await params;
  const t = await getTranslations("skills");

  let skill;
  try {
    skill = await getSkill(id);
  } catch {
    notFound();
  }

  const versionsData = await getSkillVersions().catch(() => ({ items: [] as Array<{ id: string; skill_id: string; version: string; status: string; change_summary: string; approval_required: boolean; created_at: string }> }));
  const versions = versionsData.items.filter((v) => v.skill_id === id);

  const specData = await getSkillSpec(id).catch(() => ({ spec_yaml: "" }));

  return (
    <>
      <Breadcrumb />
      <PageHeader
        eyebrow={t("eyebrow")}
        title={skill.name}
        description={
          <>
            {skill.owner_team} · {skill.risk_level} risk ·{" "}
            <span className={`badge badge-${skill.status}`}>{skill.status}</span>
          </>
        }
      />

      <section className="panel-grid panel-grid-wide">
        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">Skill</p>
            <h2>{t("name")}</h2>
          </div>
          <div className="detail-grid">
            <div className="detail-row">
              <span className="detail-label">ID</span>
              <span className="detail-value">{skill.id}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("name")}</span>
              <span className="detail-value">{skill.name}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("ownerTeam")}</span>
              <span className="detail-value">{skill.owner_team}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("riskLevel")}</span>
              <span className="detail-value">{skill.risk_level}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("status")}</span>
              <span className="detail-value">{skill.status}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">{t("version")}</span>
              <span className="detail-value">{skill.current_version || "—"}</span>
            </div>
          </div>
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">{t("versions")}</p>
            <h2>{t("versions")} ({versions.length})</h2>
          </div>
          <div className="list">
            {versions.length === 0 ? (
              <p className="empty-note">{t("noVersions")}</p>
            ) : (
              versions.map((version) => (
                <article className="list-card" key={version.id}>
                  <div>
                    <h3>
                      {version.version} · {version.status}
                    </h3>
                    <p>{version.change_summary}</p>
                    <p className="meta">
                      {version.approval_required ? t("approvalRequired") : t("noApproval")} ·{" "}
                      {new Date(version.created_at).toLocaleDateString()}
                    </p>
                  </div>
                  <span className={`badge badge-${version.status}`}>{version.status}</span>
                </article>
              ))
            )}
          </div>
        </article>

        <article className="panel">
          <div className="panel-header">
            <p className="section-kicker">{t("yamlSpec")}</p>
            <h2>{t("yamlSpec")}</h2>
          </div>
          <pre className="json-block">{specData.spec_yaml}</pre>
        </article>
      </section>
    </>
  );
}
