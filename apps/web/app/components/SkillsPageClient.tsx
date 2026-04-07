"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { Skill, createSkill } from "../lib/api";
import { useToast } from "../../components/ToastProvider";

type Props = {
  skills: { items: Skill[] };
};

export default function SkillsPageClient({ skills }: Props) {
  const t = useTranslations("skills");
  const tc = useTranslations("common");
  const router = useRouter();
  const toast = useToast();
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleCreateSkill(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setError("");

    const formData = new FormData(e.currentTarget);
    const name = formData.get("name") as string;
    const ownerTeam = formData.get("owner_team") as string;
    const riskLevel = formData.get("risk_level") as string;

    try {
      await createSkill({ name, owner_team: ownerTeam, risk_level: riskLevel });
      toast.success(tc("success") + ": Skill created");
      setShowCreateModal(false);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create skill");
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <header className="hero-page hero-compact" style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
        <div>
          <p className="eyebrow">{t("eyebrow")}</p>
          <h1>{t("title")}</h1>
          <p className="lede">{t("lede")}</p>
        </div>
        <button
          className="btn btn-primary"
          onClick={() => setShowCreateModal(true)}
          style={{ flexShrink: 0 }}
        >
          + {t("registerSkill")}
        </button>
      </header>

      <div className="panel">
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
      </div>

      {showCreateModal && (
        <div
          style={{
            position: "fixed",
            inset: 0,
            background: "rgba(0, 0, 0, 0.5)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 1000,
            padding: "var(--space-6)",
          }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setShowCreateModal(false);
          }}
        >
          <div
            style={{
              background: "var(--surface-raised)",
              borderRadius: "var(--radius-lg)",
              padding: "var(--space-6)",
              width: "100%",
              maxWidth: 480,
              maxHeight: "90vh",
              overflowY: "auto",
              boxShadow: "var(--shadow-lg)",
            }}
          >
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "var(--space-5)" }}>
              <div>
                <p className="section-kicker">{t("eyebrow")}</p>
                <h2 style={{ fontSize: "var(--text-xl)", fontWeight: 700, color: "var(--ink-strong)" }}>
                  {t("registerSkill")}
                </h2>
              </div>
              <button
                onClick={() => setShowCreateModal(false)}
                style={{
                  background: "none",
                  border: "none",
                  fontSize: "var(--text-xl)",
                  cursor: "pointer",
                  color: "var(--muted)",
                  padding: "var(--space-2)",
                }}
              >
                ×
              </button>
            </div>

            {error && <p className="form-error" style={{ marginBottom: "var(--space-4)" }}>{error}</p>}

            <form onSubmit={handleCreateSkill}>
              <div className="form-fields">
                <label className="form-label">
                  {t("skillName")}
                  <input
                    type="text"
                    name="name"
                    placeholder={t("skillNamePlaceholder")}
                    required
                    className="form-input"
                  />
                </label>

                <label className="form-label">
                  {t("ownerTeam")}
                  <input
                    type="text"
                    name="owner_team"
                    placeholder={t("ownerTeamPlaceholder")}
                    required
                    className="form-input"
                  />
                </label>

                <label className="form-label">
                  {t("riskLevel")}
                  <select name="risk_level" className="form-select" defaultValue="medium">
                    <option value="low">{t("low")}</option>
                    <option value="medium">{t("medium")}</option>
                    <option value="high">{t("high")}</option>
                  </select>
                </label>
              </div>

              <button type="submit" disabled={loading} className="form-submit" style={{ marginTop: "var(--space-4)" }}>
                {loading ? t("registering") : t("registerSkill")}
              </button>
            </form>
          </div>
        </div>
      )}
    </>
  );
}
