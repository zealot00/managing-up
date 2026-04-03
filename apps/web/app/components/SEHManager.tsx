"use client";

import { useState } from "react";
import CreateDatasetForm from "./CreateDatasetForm";
import CreatePolicyForm from "./CreatePolicyForm";
import { useTranslations } from "next-intl";

type Dataset = { dataset_id: string; name: string; version: string; owner: string; case_count: number };
type Run = { run_id: string; skill: string; metrics: { score: number; success_rate: number } };
type Policy = { policy_id: string; name: string; require_provenance: boolean; min_source_diversity: number; min_golden_weight: number };

type Props = {
  summary: { total_datasets: number; total_runs: number; total_policies: number; avg_score: number };
  datasets: Dataset[];
  runs: Run[];
  policies: Policy[];
};

export default function SEHManager({ summary, datasets, runs, policies }: Props) {
  const t = useTranslations("seh");
  const tc = useTranslations("common");
  const [activeTab, setActiveTab] = useState<"datasets" | "runs" | "policies">("datasets");
  const [showCreateDataset, setShowCreateDataset] = useState(false);
  const [showCreatePolicy, setShowCreatePolicy] = useState(false);

  const metrics = [
    { label: t("datasets"), value: summary.total_datasets, icon: "□" },
    { label: t("runs"), value: summary.total_runs, icon: "▶" },
    { label: t("policies"), value: summary.total_policies, icon: "◎" },
    { label: t("avgSkillScore").split(" ").slice(1).join(" "), value: summary.avg_score > 0 ? summary.avg_score.toFixed(2) : "—", icon: "◉" },
  ];

  const tabs: { key: "datasets" | "runs" | "policies"; label: string; count: number }[] = [
    { key: "datasets", label: t("datasets"), count: datasets.length },
    { key: "runs", label: t("runs"), count: runs.length },
    { key: "policies", label: t("policies"), count: policies.length },
  ];

  return (
    <>
      <div className="stats">
        {metrics.map((metric) => (
          <article className="metric-card" key={metric.label}>
            <div className="metric-card-icon">{metric.icon}</div>
            <div className="metric-card-value">{metric.value}</div>
            <div className="metric-card-label">{metric.label}</div>
          </article>
        ))}
      </div>

      <div className="page-header" style={{ marginTop: "var(--space-6)", marginBottom: "var(--space-4)", paddingBottom: 0, borderBottom: "none" }}>
        <div className="page-header-content">
          <div style={{ display: "flex", gap: "var(--space-2)" }}>
            {tabs.map((tab) => (
              <button
                key={tab.key}
                className={`btn btn-sm ${activeTab === tab.key ? "btn-primary" : "btn-secondary"}`}
                onClick={() => {
                  setActiveTab(tab.key);
                  setShowCreateDataset(false);
                  setShowCreatePolicy(false);
                }}
              >
                {tab.label} ({tab.count})
              </button>
            ))}
          </div>
        </div>
        <div className="page-header-actions">
          {activeTab === "datasets" && (
            <button className="btn btn-primary" onClick={() => { setShowCreateDataset(!showCreateDataset); setShowCreatePolicy(false); }}>
              {showCreateDataset ? tc("cancel") : t("newDataset")}
            </button>
          )}
          {activeTab === "policies" && (
            <button className="btn btn-primary" onClick={() => { setShowCreatePolicy(!showCreatePolicy); setShowCreateDataset(false); }}>
              {showCreatePolicy ? tc("cancel") : t("newPolicy")}
            </button>
          )}
        </div>
      </div>

      {activeTab === "datasets" && showCreateDataset && (
        <CreateDatasetForm onCreated={() => setShowCreateDataset(false)} />
      )}

      {activeTab === "policies" && showCreatePolicy && (
        <CreatePolicyForm onCreated={() => setShowCreatePolicy(false)} />
      )}

      <section className="panel" style={{ marginTop: "var(--space-4)" }}>
        {activeTab === "datasets" && (
          <>
            <div className="panel-header">
              <p className="section-kicker">{t("datasets")}</p>
              <h2 className="panel-title">{t("allDatasets", { count: datasets.length })}</h2>
            </div>
            <div className="list">
              {datasets.length === 0 ? (
                <p className="empty-note">{t("noDatasets")}</p>
              ) : (
                datasets.map((ds) => (
                  <article className="list-card" key={ds.dataset_id}>
                    <div className="list-card-main">
                      <h3 className="list-card-title">{ds.name}</h3>
                      <p className="list-card-meta">
                        {ds.version} · {ds.owner} · {ds.case_count} {t("cases")}
                      </p>
                    </div>
                    <span className="badge badge-muted">{ds.dataset_id}</span>
                  </article>
                ))
              )}
            </div>
          </>
        )}

        {activeTab === "runs" && (
          <>
            <div className="panel-header">
              <p className="section-kicker">{t("performance")}</p>
              <h2 className="panel-title">{t("allRuns", { count: runs.length })}</h2>
            </div>
            <div className="list">
              {runs.length === 0 ? (
                <p className="empty-note">{t("noRuns")}</p>
              ) : (
                runs.map((run) => (
                  <article className="list-card" key={run.run_id}>
                    <div className="list-card-main">
                      <h3 className="list-card-title">{run.skill}</h3>
                      <p className="list-card-meta">
                        {t("score")}: {run.metrics.score.toFixed(2)} · {t("success")}: {(run.metrics.success_rate * 100).toFixed(0)}%
                      </p>
                    </div>
                    <span className={`badge badge-${run.metrics.score >= 0.75 ? "succeeded" : "failed"}`}>
                      {run.run_id}
                    </span>
                  </article>
                ))
              )}
            </div>
          </>
        )}

        {activeTab === "policies" && (
          <>
            <div className="panel-header">
              <p className="section-kicker">{t("governance")}</p>
              <h2 className="panel-title">{t("activePolicies", { count: policies.length })}</h2>
            </div>
            <div className="list">
              {policies.length === 0 ? (
                <p className="empty-note">{t("noPolicies")}</p>
              ) : (
                policies.map((policy) => (
                  <article className="list-card" key={policy.policy_id}>
                    <div className="list-card-main">
                      <h3 className="list-card-title">{policy.name}</h3>
                      <p className="list-card-meta">
                        {policy.require_provenance ? `${t("provenanceRequired")} · ` : ""}
                        {t("minDiversity")}: {policy.min_source_diversity} · {t("minGolden")}: {policy.min_golden_weight}
                      </p>
                    </div>
                    <span className="badge badge-muted">{policy.policy_id}</span>
                  </article>
                ))
              )}
            </div>
          </>
        )}
      </section>
    </>
  );
}