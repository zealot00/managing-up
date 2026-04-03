"use client";

import Link from "next/link";
import { useState, useMemo } from "react";
import { useRouter } from "next/navigation";
import CreateDatasetForm from "./CreateDatasetForm";
import CreatePolicyForm from "./CreatePolicyForm";
import EditPolicyForm from "./EditPolicyForm";
import { useTranslations } from "next-intl";
import { deleteSEHDataset, createSEHRelease } from "../lib/seh-api";

type Dataset = { dataset_id: string; name: string; version: string; owner: string; case_count: number };
type Run = { run_id: string; skill: string; dataset_id: string; metrics: { score: number; success_rate: number } };
type Policy = { policy_id: string; name: string; require_provenance: boolean; min_source_diversity: number; min_golden_weight: number; source_policies?: unknown[] };

type Props = {
  summary: { total_datasets: number; total_runs: number; total_policies: number; avg_score: number };
  datasets: Dataset[];
  runs: Run[];
  policies: Policy[];
};

export default function SEHManager({ summary, datasets, runs, policies }: Props) {
  const t = useTranslations("seh");
  const tc = useTranslations("common");
  const router = useRouter();
  const [activeTab, setActiveTab] = useState<"datasets" | "runs" | "policies">("datasets");
  const [showCreateDataset, setShowCreateDataset] = useState(false);
  const [showCreatePolicy, setShowCreatePolicy] = useState(false);
  const [editingPolicy, setEditingPolicy] = useState<Policy | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [releasingId, setReleasingId] = useState<string | null>(null);
  const [expandedCase, setExpandedCase] = useState<string | null>(null);

  const metrics = [
    { label: t("datasets"), value: summary.total_datasets, icon: "□" },
    { label: t("runs"), value: summary.total_runs, icon: "▶" },
    { label: t("policies"), value: summary.total_policies, icon: "◎" },
    { label: t("avgSkillScore"), value: summary.avg_score > 0 ? summary.avg_score.toFixed(2) : "—", icon: "◉" },
  ];

  const tabs: { key: "datasets" | "runs" | "policies"; label: string; count: number }[] = [
    { key: "datasets", label: t("datasets"), count: datasets.length },
    { key: "runs", label: t("runs"), count: runs.length },
    { key: "policies", label: t("policies"), count: policies.length },
  ];

  const filteredDatasets = useMemo(() => {
    if (!searchQuery) return datasets;
    const q = searchQuery.toLowerCase();
    return datasets.filter(d =>
      d.name.toLowerCase().includes(q) ||
      d.owner.toLowerCase().includes(q) ||
      d.dataset_id.toLowerCase().includes(q)
    );
  }, [datasets, searchQuery]);

  const filteredRuns = useMemo(() => {
    if (!searchQuery) return runs;
    const q = searchQuery.toLowerCase();
    return runs.filter(r =>
      r.skill.toLowerCase().includes(q) ||
      r.run_id.toLowerCase().includes(q)
    );
  }, [runs, searchQuery]);

  const filteredPolicies = useMemo(() => {
    if (!searchQuery) return policies;
    const q = searchQuery.toLowerCase();
    return policies.filter(p =>
      p.name.toLowerCase().includes(q) ||
      p.policy_id.toLowerCase().includes(q)
    );
  }, [policies, searchQuery]);

  async function handleDeleteDataset(id: string, name: string) {
    if (!confirm(t("confirmDelete", { name }))) return;
    setDeletingId(id);
    try {
      await deleteSEHDataset(id);
      router.refresh();
    } catch {
      alert("Failed to delete dataset");
    } finally {
      setDeletingId(null);
    }
  }

  async function handleTriggerRelease(skillId: string) {
    setReleasingId(skillId);
    try {
      await createSEHRelease(skillId);
      router.refresh();
    } catch {
      alert("Failed to create release");
    } finally {
      setReleasingId(null);
    }
  }

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
                  setEditingPolicy(null);
                  setSearchQuery("");
                }}
              >
                {tab.label} ({tab.count})
              </button>
            ))}
          </div>
        </div>
        <div className="page-header-actions">
          {activeTab === "datasets" && (
            <button className="btn btn-primary" onClick={() => { setShowCreateDataset(!showCreateDataset); setShowCreatePolicy(false); setEditingPolicy(null); }}>
              {showCreateDataset ? tc("cancel") : t("newDataset")}
            </button>
          )}
          {activeTab === "policies" && !editingPolicy && (
            <button className="btn btn-primary" onClick={() => { setShowCreatePolicy(!showCreatePolicy); setShowCreateDataset(false); }}>
              {showCreatePolicy ? tc("cancel") : t("newPolicy")}
            </button>
          )}
        </div>
      </div>

      {activeTab === "datasets" && showCreateDataset && (
        <CreateDatasetForm onCreated={() => { setShowCreateDataset(false); router.refresh(); }} />
      )}

      {activeTab === "policies" && showCreatePolicy && (
        <CreatePolicyForm onCreated={() => { setShowCreatePolicy(false); router.refresh(); }} />
      )}

      {activeTab === "policies" && editingPolicy && (
        <EditPolicyForm
          policy={editingPolicy}
          onCancel={() => setEditingPolicy(null)}
          onUpdated={() => { setEditingPolicy(null); router.refresh(); }}
        />
      )}

      {(activeTab === "datasets" || activeTab === "runs" || activeTab === "policies") && (
        <div className="search-bar" style={{ marginTop: "var(--space-2)" }}>
          <input
            type="text"
            placeholder={t("searchPlaceholder")}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
      )}

      <section className="panel" style={{ marginTop: "var(--space-4)" }}>
        {activeTab === "datasets" && (
          <>
            <div className="panel-header">
              <p className="section-kicker">{t("datasets")}</p>
              <h2 className="panel-title">{t("allDatasets", { count: filteredDatasets.length })}</h2>
              <Link href="/seh/datasets" className="btn btn-secondary btn-sm">
                {t("viewAllDatasets")}
              </Link>
            </div>
            <div className="list">
              {filteredDatasets.length === 0 ? (
                <p className="empty-note">{searchQuery ? "No matching datasets" : t("noDatasets")}</p>
              ) : (
                filteredDatasets.map((ds) => (
                  <article className="list-card" key={ds.dataset_id}>
                    <div className="list-card-main">
                      <h3 className="list-card-title">{ds.name}</h3>
                      <p className="list-card-meta">
                        {ds.version} · {ds.owner} · {ds.case_count} {t("cases")}
                      </p>
                    </div>
                    <div className="list-card-actions">
                      <span className="badge badge-muted">{ds.dataset_id}</span>
                      <Link href={`/seh/datasets/${ds.dataset_id}`} className="btn btn-secondary btn-sm">
                        {t("viewDetails")}
                      </Link>
                      <button
                        className="btn btn-sm btn-ghost"
                        onClick={() => handleDeleteDataset(ds.dataset_id, ds.name)}
                        disabled={deletingId === ds.dataset_id}
                      >
                        {deletingId === ds.dataset_id ? t("deleting") : t("delete")}
                      </button>
                    </div>
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
              <h2 className="panel-title">{t("allRuns", { count: filteredRuns.length })}</h2>
              <Link href="/seh/runs" className="btn btn-secondary btn-sm">
                {t("viewAllRuns")}
              </Link>
            </div>
            <div className="list">
              {filteredRuns.length === 0 ? (
                <p className="empty-note">{searchQuery ? "No matching runs" : t("noRuns")}</p>
              ) : (
                filteredRuns.map((run) => (
                  <article className="list-card" key={run.run_id}>
                    <div className="list-card-main">
                      <h3 className="list-card-title">{run.skill}</h3>
                      <p className="list-card-meta">
                        {t("score")}: {run.metrics.score.toFixed(2)} · {t("success")}: {(run.metrics.success_rate * 100).toFixed(0)}%
                      </p>
                    </div>
                    <div className="list-card-actions">
                      <span className={`badge badge-${run.metrics.score >= 0.75 ? "succeeded" : "failed"}`}>
                        {run.run_id}
                      </span>
                      <Link href={`/seh/runs/${run.run_id}`} className="btn btn-secondary btn-sm">
                        {t("viewDetails")}
                      </Link>
                      <button
                        className="btn btn-sm btn-ghost"
                        onClick={() => handleTriggerRelease(run.skill)}
                        disabled={releasingId === run.skill}
                      >
                        {releasingId === run.skill ? t("triggering") : t("triggerRun")}
                      </button>
                    </div>
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
              <h2 className="panel-title">{t("activePolicies", { count: filteredPolicies.length })}</h2>
              <Link href="/seh/policies" className="btn btn-secondary btn-sm">
                {t("viewAllPolicies")}
              </Link>
            </div>
            <div className="list">
              {filteredPolicies.length === 0 ? (
                <p className="empty-note">{searchQuery ? "No matching policies" : t("noPolicies")}</p>
              ) : (
                filteredPolicies.map((policy) => (
                  <article className="list-card" key={policy.policy_id}>
                    <div className="list-card-main">
                      <h3 className="list-card-title">{policy.name}</h3>
                      <p className="list-card-meta">
                        {policy.require_provenance ? `${t("provenanceRequired")} · ` : ""}
                        {t("minDiversity")}: {policy.min_source_diversity} · {t("minGolden")}: {policy.min_golden_weight}
                      </p>
                    </div>
                    <div className="list-card-actions">
                      <span className="badge badge-muted">{policy.policy_id}</span>
                      <button
                        className="btn btn-sm btn-secondary"
                        onClick={() => setEditingPolicy(policy)}
                      >
                        {tc("edit")}
                      </button>
                    </div>
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
