"use client";

import Link from "next/link";
import { useState, useMemo } from "react";
import CreateDatasetForm from "./CreateDatasetForm";
import CreatePolicyForm from "./CreatePolicyForm";
import EditPolicyForm from "./EditPolicyForm";
import { useTranslations } from "next-intl";
import { deleteSEHDataset, createSEHRelease } from "../lib/seh-api";
import { useApiMutation } from "../lib/use-mutations";
import { ConfirmDialog } from "./ui/ConfirmDialog";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";
import { Database, Play, Target, LayoutDashboard } from "lucide-react";

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
  const [activeTab, setActiveTab] = useState<"datasets" | "runs" | "policies">("datasets");
  const [showCreateDataset, setShowCreateDataset] = useState(false);
  const [showCreatePolicy, setShowCreatePolicy] = useState(false);
  const [editingPolicy, setEditingPolicy] = useState<Policy | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [expandedCase, setExpandedCase] = useState<string | null>(null);
  const [deleteDialogData, setDeleteDialogData] = useState<{ id: string; name: string } | null>(null);

  const deleteDatasetMutation = useApiMutation(deleteSEHDataset, {
    queryKeysToInvalidate: [["datasets"]],
    successMessage: tc("success") + ": Dataset deleted",
    onSuccess: () => {
      setDeleteDialogData(null);
    },
  });

  const createReleaseMutation = useApiMutation(createSEHRelease, {
    queryKeysToInvalidate: [["releases"]],
    successMessage: tc("success") + ": Release created",
  });

  const metrics = [
    { label: t("datasets"), value: summary.total_datasets, icon: <Database size={20} aria-hidden="true" /> },
    { label: t("runs"), value: summary.total_runs, icon: <Play size={20} aria-hidden="true" /> },
    { label: t("policies"), value: summary.total_policies, icon: <Target size={20} aria-hidden="true" /> },
    { label: t("avgSkillScore"), value: summary.avg_score > 0 ? summary.avg_score.toFixed(2) : "—", icon: <LayoutDashboard size={20} aria-hidden="true" /> },
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

  function handleDeleteDataset() {
    if (!deleteDialogData) return;
    deleteDatasetMutation.mutate(deleteDialogData.id);
  }

  function handleTriggerRelease(skillId: string) {
    createReleaseMutation.mutate(skillId);
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

      <PageHeader
        eyebrow={
          <div className="tabs">
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
        }
        title=""
        actions={
          <>
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
          </>
        }
      />

      {activeTab === "datasets" && showCreateDataset && (
        <CreateDatasetForm onCreated={() => { setShowCreateDataset(false); }} />
      )}

      {activeTab === "policies" && showCreatePolicy && (
        <CreatePolicyForm onCreated={() => { setShowCreatePolicy(false); }} />
      )}

      {activeTab === "policies" && editingPolicy && (
        <EditPolicyForm
          policy={editingPolicy}
          onCancel={() => setEditingPolicy(null)}
          onUpdated={() => { setEditingPolicy(null); }}
        />
      )}

      {(activeTab === "datasets" || activeTab === "runs" || activeTab === "policies") && (
        <div className="search-bar">
          <input
            type="text"
            placeholder={t("searchPlaceholder")}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
      )}

      <section className="panel">
        {activeTab === "datasets" && (
          <>
            <div className="panel-header">
              <p className="section-kicker">{t("datasets")}</p>
              <h2 className="panel-title">{t("allDatasets", { count: filteredDatasets.length })}</h2>
              <Link href="/seh/datasets" className="btn btn-secondary btn-sm">
                {t("viewAllDatasets")}
              </Link>
            </div>
            <div className="table-wrapper">
              {filteredDatasets.length === 0 ? (
                <EmptyState
                  title={searchQuery ? "No matching datasets" : t("noDatasets")}
                />
              ) : (
                <table className="table">
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Version</th>
                      <th>Owner</th>
                      <th>Cases</th>
                      <th>Dataset ID</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredDatasets.map((ds) => (
                      <tr key={ds.dataset_id}>
                        <td>{ds.name}</td>
                        <td>{ds.version}</td>
                        <td>{ds.owner}</td>
                        <td>{ds.case_count}</td>
                        <td><span className="badge badge-muted">{ds.dataset_id}</span></td>
                        <td>
                          <div className="flex gap-2">
                            <Link href={`/seh/datasets/${ds.dataset_id}`} className="btn btn-secondary btn-sm">
                              {t("viewDetails")}
                            </Link>
                            <button
                              className="btn btn-sm btn-ghost"
                              onClick={() => setDeleteDialogData({ id: ds.dataset_id, name: ds.name })}
                              disabled={deleteDatasetMutation.isPending}
                            >
                              {deleteDatasetMutation.isPending ? t("deleting") : t("delete")}
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
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
            <div className="table-wrapper">
              {filteredRuns.length === 0 ? (
                <p className="empty-note">{searchQuery ? "No matching runs" : t("noRuns")}</p>
              ) : (
                <table className="table">
                  <thead>
                    <tr>
                      <th>Skill</th>
                      <th>Dataset</th>
                      <th>Score</th>
                      <th>Success Rate</th>
                      <th>Run ID</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredRuns.map((run) => (
                      <tr key={run.run_id}>
                        <td>{run.skill}</td>
                        <td>{run.dataset_id}</td>
                        <td>
                          <span className={`badge badge-${run.metrics.score >= 0.75 ? "succeeded" : "failed"}`}>
                            {run.metrics.score.toFixed(2)}
                          </span>
                        </td>
                        <td>{(run.metrics.success_rate * 100).toFixed(0)}%</td>
                        <td><span className="badge badge-muted">{run.run_id}</span></td>
                        <td>
                          <div className="flex gap-2">
                            <Link href={`/seh/runs/${run.run_id}`} className="btn btn-secondary btn-sm">
                              {t("viewDetails")}
                            </Link>
                            <button
                              className="btn btn-sm btn-ghost"
                              onClick={() => handleTriggerRelease(run.skill)}
                              disabled={createReleaseMutation.isPending}
                            >
                              {createReleaseMutation.isPending ? t("triggering") : t("triggerRun")}
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
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
            <div className="table-wrapper">
              {filteredPolicies.length === 0 ? (
                <p className="empty-note">{searchQuery ? "No matching policies" : t("noPolicies")}</p>
              ) : (
                <table className="table">
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Provenance</th>
                      <th>Min Diversity</th>
                      <th>Min Golden</th>
                      <th>Policy ID</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredPolicies.map((policy) => (
                      <tr key={policy.policy_id}>
                        <td>{policy.name}</td>
                        <td>{policy.require_provenance ? t("provenanceRequired") : "—"}</td>
                        <td>{policy.min_source_diversity}</td>
                        <td>{policy.min_golden_weight}</td>
                        <td><span className="badge badge-muted">{policy.policy_id}</span></td>
                        <td>
                          <button
                            className="btn btn-sm btn-secondary"
                            onClick={() => setEditingPolicy(policy)}
                          >
                            {tc("edit")}
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </>
        )}
      </section>

      <ConfirmDialog
        isOpen={deleteDialogData !== null}
        onClose={() => setDeleteDialogData(null)}
        onConfirm={handleDeleteDataset}
        title={tc("deleteConfirmTitle", { name: deleteDialogData?.name || "" })}
        description={tc("deleteConfirmDescription")}
        confirmText={t("delete")}
        cancelText={tc("cancel")}
        variant="danger"
      />
    </>
  );
}
