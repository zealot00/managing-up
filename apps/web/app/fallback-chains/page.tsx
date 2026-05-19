"use client";

import { FormEvent, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { useToast } from "../../components/ToastProvider";
import { ConfirmDialog } from "../components/ui/ConfirmDialog";
import Breadcrumb from "../../components/Breadcrumb";
import {
  FallbackChain,
  listFallbackChains,
  createFallbackChain,
  updateFallbackChain,
  deleteFallbackChain,
} from "../lib/fallback-api";
import {
  ArrowDownUp,
  CheckCircle,
  ChevronDown,
  ChevronRight,
  XCircle,
  Plus,
  Trash2,
  X,
  Power,
  PowerOff,
} from "lucide-react";
import { Spinner } from "../components/ui/Spinner";

const PROVIDERS = [
  "openai",
  "anthropic",
  "google",
  "azure",
  "deepseek",
  "zhipuai",
  "baidu",
  "alibaba",
  "minimax",
  "ollama",
];

type TargetInput = {
  provider: string;
  model: string;
  weight: number;
  priority: number;
  is_enabled: boolean;
};

export default function FallbackChainsPage() {
  const t = useTranslations("fallbackChains");
  const tc = useTranslations("common");
  const toast = useToast();

  const [chains, setChains] = useState<FallbackChain[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [expandedChainId, setExpandedChainId] = useState<string | null>(null);

  const [formModel, setFormModel] = useState("");
  const [formTargets, setFormTargets] = useState<TargetInput[]>([
    { provider: "openai", model: "", weight: 1, priority: 1, is_enabled: true },
  ]);

  async function loadData() {
    setError(null);
    setIsLoading(true);
    try {
      const resp = await listFallbackChains();
      setChains(resp.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load fallback chains");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void loadData();
  }, []);

  function resetForm() {
    setFormModel("");
    setFormTargets([
      { provider: "openai", model: "", weight: 1, priority: 1, is_enabled: true },
    ]);
    setShowCreateForm(false);
  }

  function addTargetRow() {
    setFormTargets((prev) => [
      ...prev,
      { provider: "openai", model: "", weight: 1, priority: prev.length + 1, is_enabled: true },
    ]);
  }

  function removeTargetRow(index: number) {
    setFormTargets((prev) => prev.filter((_, i) => i !== index));
  }

  function updateTargetRow(index: number, field: keyof TargetInput, value: string | number | boolean) {
    setFormTargets((prev) =>
      prev.map((row, i) => (i === index ? { ...row, [field]: value } : row))
    );
  }

  async function handleCreate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!formModel.trim()) return;

    setIsSubmitting(true);
    setError(null);
    try {
      await createFallbackChain({
        model: formModel.trim(),
        is_enabled: true,
        targets: formTargets.filter((t) => t.model.trim()).map((t) => ({
          provider: t.provider,
          model: t.model.trim(),
          weight: t.weight,
          priority: t.priority,
          is_enabled: t.is_enabled,
        })),
      });
      toast.success(tc("success") + ": " + t("chainCreated"));
      resetForm();
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("createFailed"));
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteFallbackChain(id);
      toast.success(tc("success") + ": " + t("chainDeleted"));
      setDeletingId(null);
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("deleteFailed"));
    }
  }

  async function handleToggleEnabled(chain: FallbackChain) {
    try {
      await updateFallbackChain(chain.id, {
        is_enabled: !chain.is_enabled,
      });
      await loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("updateFailed"));
    }
  }

  function toggleExpanded(chainId: string) {
    setExpandedChainId((prev) => (prev === chainId ? null : chainId));
  }

  if (isLoading) {
    return (
      <div className="admin-content">
        <Breadcrumb />
        <div className="loading-pulse loading-pulse-long" style={{ marginBottom: 16 }} />
        <div className="skeleton-grid">
          <div className="skeleton-card" />
          <div className="skeleton-card" />
          <div className="skeleton-card" />
        </div>
      </div>
    );
  }

  return (
    <div className="admin-content">
      <Breadcrumb />

      <div className="page-header">
        <div className="page-header-content">
          <p className="section-kicker">{t("title")}</p>
          <h1 className="panel-title">{t("subtitle")}</h1>
        </div>
        <div className="page-header-actions">
          <button
            className="btn btn-primary"
            onClick={() => setShowCreateForm(true)}
          >
            <Plus size={16} aria-hidden="true" />
            {t("newChain")}
          </button>
        </div>
      </div>

      {error && (
        <div className="form-error" style={{ marginBottom: 16 }} role="alert">
          {error}
        </div>
      )}

      {showCreateForm && (
        <div className="form-panel">
          <div className="form-header" style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
            <div>
              <p className="section-kicker">{t("title")}</p>
              <h2 className="form-title">{t("createChain")}</h2>
            </div>
            <button
              className="btn btn-ghost btn-sm"
              onClick={resetForm}
              aria-label={tc("close")}
            >
              <X size={18} aria-hidden="true" />
            </button>
          </div>
          <form className="form-fields" onSubmit={handleCreate}>
            <div>
              <label className="form-label" htmlFor="fc-model">{t("model")}</label>
              <input
                id="fc-model"
                className="form-input"
                value={formModel}
                onChange={(e) => setFormModel(e.target.value)}
                placeholder={t("modelPlaceholder")}
                required
                autoFocus
              />
            </div>

            <div>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
                <label className="form-label" id="fc-targets-label">{t("targets")}</label>
                <button
                  type="button"
                  className="btn btn-sm btn-secondary"
                  onClick={addTargetRow}
                >
                  <Plus size={14} aria-hidden="true" />
                  {t("addTarget")}
                </button>
              </div>
              {formTargets.map((target, index) => (
                <div key={index} style={{ display: "flex", gap: 8, marginBottom: 8, alignItems: "center" }}>
                  <select
                    className="form-select"
                    value={target.provider}
                    onChange={(e) => updateTargetRow(index, "provider", e.target.value)}
                    style={{ minWidth: 120 }}
                    aria-label={`${t("provider")} ${index + 1}`}
                  >
                    {PROVIDERS.map((p) => (
                      <option key={p} value={p}>{p}</option>
                    ))}
                  </select>
                  <input
                    className="form-input"
                    value={target.model}
                    onChange={(e) => updateTargetRow(index, "model", e.target.value)}
                    placeholder={t("targetModelPlaceholder")}
                    style={{ flex: 1 }}
                    aria-label={`${t("model")} ${index + 1}`}
                  />
                  <input
                    className="form-input"
                    type="number"
                    value={target.priority}
                    onChange={(e) => updateTargetRow(index, "priority", Number(e.target.value))}
                    placeholder={t("priority")}
                    style={{ width: 80 }}
                    min={1}
                    aria-label={`${t("priority")} ${index + 1}`}
                  />
                  {formTargets.length > 1 && (
                    <button
                      type="button"
                      className="btn btn-ghost btn-sm"
                      onClick={() => removeTargetRow(index)}
                      aria-label={tc("delete")}
                    >
                      <Trash2 size={14} aria-hidden="true" />
                    </button>
                  )}
                </div>
              ))}
            </div>

            <div className="form-actions">
              <button type="button" className="btn btn-secondary" onClick={resetForm}>
                {tc("cancel")}
              </button>
              <button type="submit" className="btn btn-primary" style={{ display: "flex", alignItems: "center", gap: 6 }}>
                {isSubmitting ? <><Spinner size="sm" /> {t("creating")}</> : t("create")}
              </button>
            </div>
          </form>
        </div>
      )}

      {chains.length === 0 ? (
        <div className="empty-state">
          <div className="empty-state-icon">
            <ArrowDownUp size={48} aria-hidden="true" />
          </div>
          <h3 className="empty-state-title">{t("noChains")}</h3>
          <p className="empty-state-description">
            {t("noChainsDesc")}
          </p>
          <div className="empty-state-action" style={{ marginTop: 16 }}>
            <button className="btn btn-primary" onClick={() => setShowCreateForm(true)}>
              <Plus size={16} aria-hidden="true" />
              {t("newChain")}
            </button>
          </div>
        </div>
      ) : (
        <div className="panel">
          <div className="gateway-table-wrapper">
            <table className="gateway-table">
              <thead>
                <tr>
                  <th scope="col"></th>
                  <th>{t("model")}</th>
                  <th style={{ width: 100 }}>{tc("status")}</th>
                  <th style={{ width: 100 }}>{t("targetCount")}</th>
                  <th style={{ width: 80 }}>{t("enabled")}</th>
                  <th style={{ width: 120 }}>{tc("actions")}</th>
                </tr>
              </thead>
              <tbody>
                {chains.map((chain) => {
                  const isDeleting = deletingId === chain.id;
                  const isExpanded = expandedChainId === chain.id;
                  const sortedTargets = [...chain.targets].sort((a, b) => a.priority - b.priority);

                  return (
                    <>
                      <tr key={chain.id}>
                        <td>
                          <button
                            className="btn btn-ghost btn-sm"
                            onClick={() => toggleExpanded(chain.id)}
                            aria-label={isExpanded ? "Collapse" : "Expand"}
                            aria-expanded={isExpanded}
                          >
                            {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                          </button>
                        </td>
                        <td>
                          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                            <ArrowDownUp size={16} aria-hidden="true" style={{ color: "var(--muted)", flexShrink: 0 }} />
                            <div>
                              <div style={{ fontWeight: 600, color: "var(--ink-strong)", fontSize: "var(--text-sm)" }}>
                                {chain.model}
                              </div>
                            </div>
                          </div>
                        </td>
                        <td>
                          <span className={chain.is_enabled ? "badge badge-completed" : "badge badge-pending"} style={{ display: "inline-flex", alignItems: "center", gap: 4 }}>
                            {chain.is_enabled ? <CheckCircle size={12} aria-hidden="true" /> : <XCircle size={12} aria-hidden="true" />}
                            {chain.is_enabled ? t("enabled") : t("disabled")}
                          </span>
                        </td>
                        <td>
                          <span style={{ fontSize: "var(--text-sm)" }}>
                            {chain.targets.length}
                          </span>
                        </td>
                        <td>
                          <button
                            className={`mcp-toggle-btn ${chain.is_enabled ? "mcp-toggle-btn-on" : "mcp-toggle-btn-off"}`}
                            onClick={() => void handleToggleEnabled(chain)}
                            title={chain.is_enabled ? tc("disable") || "Disable" : tc("enable") || "Enable"}
                            aria-label={chain.is_enabled ? "Disable chain" : "Enable chain"}
                          >
                            {chain.is_enabled ? <Power size={14} aria-hidden="true" /> : <PowerOff size={14} aria-hidden="true" />}
                          </button>
                        </td>
                        <td>
                          <button
                            className="btn btn-sm btn-ghost"
                            onClick={() => setDeletingId(chain.id)}
                            aria-label={`Delete ${chain.model}`}
                          >
                            <Trash2 size={14} aria-hidden="true" />
                          </button>
                        </td>
                      </tr>
                      {isExpanded && (
                        <tr key={`${chain.id}-expanded`}>
                          <td colSpan={6} style={{ padding: "16px", background: "var(--background-subtle)" }}>
                            <h4 style={{ fontSize: "var(--text-sm)", fontWeight: 600, marginBottom: 8 }}>
                              {t("targets")} ({sortedTargets.length})
                            </h4>
                            {sortedTargets.length === 0 ? (
                              <p style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>{t("noTargets")}</p>
                            ) : (
                              <table className="gateway-table" style={{ fontSize: "var(--text-xs)" }}>
                                <thead>
                                  <tr>
                                    <th scope="col">{t("priority")}</th>
                                    <th scope="col">{t("provider")}</th>
                                    <th scope="col">{t("model")}</th>
                                    <th scope="col">{t("weight")}</th>
                                    <th scope="col">{tc("status")}</th>
                                  </tr>
                                </thead>
                                <tbody>
                                  {sortedTargets.map((target) => (
                                    <tr key={target.id}>
                                      <td>{target.priority}</td>
                                      <td>
                                        <code style={{ fontFamily: "'IBM Plex Mono', monospace" }}>
                                          {target.provider}
                                        </code>
                                      </td>
                                      <td>
                                        <code style={{ fontFamily: "'IBM Plex Mono', monospace" }}>
                                          {target.model}
                                        </code>
                                      </td>
                                      <td>{target.weight}</td>
                                      <td>
                                        <span className={target.is_enabled ? "badge badge-completed" : "badge badge-pending"} style={{ display: "inline-flex", alignItems: "center", gap: 4, fontSize: "var(--text-xs)" }}>
                                          {target.is_enabled ? t("enabled") : t("disabled")}
                                        </span>
                                      </td>
                                    </tr>
                                  ))}
                                </tbody>
                              </table>
                            )}
                          </td>
                        </tr>
                      )}
                    </>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      <ConfirmDialog
        isOpen={deletingId !== null}
        onClose={() => setDeletingId(null)}
        onConfirm={() => deletingId && handleDelete(deletingId)}
        title={tc("deleteConfirmTitle", { name: chains.find(c => c.id === deletingId)?.model || "" })}
        description={tc("deleteConfirmDescription")}
        confirmText={tc("delete")}
        cancelText={tc("cancel")}
        variant="danger"
      />
    </div>
  );
}
