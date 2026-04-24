"use client";

import { useState, FormEvent } from "react";
import { useTranslations } from "next-intl";
import { useToast } from "../../components/ToastProvider";
import Breadcrumb from "../../components/Breadcrumb";
import {
  SweepConfig,
  SweepMatrixCell,
  SweepParameters,
  SweepPromptVariant,
  getSweeps,
  getSweep,
  createSweep,
  deleteSweep,
  getSweepMatrix,
} from "../lib/api";
import {
  Crosshair,
  Plus,
  X,
  ChevronDown,
  ChevronRight,
  Trash2,
  Play,
  Pause,
  CheckCircle2,
  AlertCircle,
  Loader2,
} from "lucide-react";

const DEFAULT_MODELS = ["gpt-4o", "gpt-4o-mini", "claude-3-5-sonnet", "claude-3-haiku"];
const DEFAULT_TEMPERATURES = [0.0, 0.3, 0.7, 1.0];
const DEFAULT_MAX_TOKENS = [256, 512, 1024, 2048];

export default function SweepsPageClient() {
  const t = useTranslations("sweeps");
  const tc = useTranslations("common");
  const toast = useToast();

  const [sweeps, setSweeps] = useState<SweepConfig[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [selectedSweep, setSelectedSweep] = useState<SweepConfig | null>(null);
  const [sweepMatrix, setSweepMatrix] = useState<SweepMatrixCell[][] | null>(null);
  const [matrixSummary, setMatrixSummary] = useState<{
    total: number;
    completed: number;
    pending: number;
    avg_score: number;
    max_score: number;
    min_score: number;
  } | null>(null);

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [formName, setFormName] = useState("");
  const [formDescription, setFormDescription] = useState("");
  const [formTaskID, setFormTaskID] = useState("");
  const [formModels, setFormModels] = useState<string[]>([DEFAULT_MODELS[0]]);
  const [formTemperatures, setFormTemperatures] = useState<number[]>([0.0]);
  const [formMaxTokens, setFormMaxTokens] = useState<number[]>([512]);
  const [formPrompts, setFormPrompts] = useState<SweepPromptVariant[]>([
    { id: "prompt_1", label: "Default", content: "" },
  ]);

  async function loadSweeps() {
    setError(null);
    setIsLoading(true);
    try {
      const resp = await getSweeps();
      setSweeps(resp.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load sweeps");
    } finally {
      setIsLoading(false);
    }
  }

  useState(() => {
    void loadSweeps();
  });

  async function loadSweepDetail(id: string) {
    try {
      const sweep = await getSweep(id);
      setSelectedSweep(sweep);
      const matrixData = await getSweepMatrix(id);
      setSweepMatrix(matrixData.matrix);
      setMatrixSummary(matrixData.summary);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to load sweep detail");
    }
  }

  function toggleExpanded(id: string) {
    if (expandedId === id) {
      setExpandedId(null);
      setSelectedSweep(null);
      setSweepMatrix(null);
      setMatrixSummary(null);
    } else {
      setExpandedId(id);
      void loadSweepDetail(id);
    }
  }

  function resetForm() {
    setFormName("");
    setFormDescription("");
    setFormTaskID("");
    setFormModels([DEFAULT_MODELS[0]]);
    setFormTemperatures([0.0]);
    setFormMaxTokens([512]);
    setFormPrompts([{ id: "prompt_1", label: "Default", content: "" }]);
    setShowCreateForm(false);
  }

  function addModel() {
    setFormModels([...formModels, DEFAULT_MODELS[0]]);
  }

  function updateModel(index: number, value: string) {
    const newModels = [...formModels];
    newModels[index] = value;
    setFormModels(newModels);
  }

  function removeModel(index: number) {
    setFormModels(formModels.filter((_, i) => i !== index));
  }

  function addTemperature() {
    setFormTemperatures([...formTemperatures, 0.7]);
  }

  function updateTemperature(index: number, value: number) {
    const newTemps = [...formTemperatures];
    newTemps[index] = value;
    setFormTemperatures(newTemps);
  }

  function removeTemperature(index: number) {
    setFormTemperatures(formTemperatures.filter((_, i) => i !== index));
  }

  function addMaxTokens() {
    setFormMaxTokens([...formMaxTokens, 1024]);
  }

  function updateMaxTokens(index: number, value: number) {
    const newTokens = [...formMaxTokens];
    newTokens[index] = value;
    setFormMaxTokens(newTokens);
  }

  function removeMaxTokens(index: number) {
    setFormMaxTokens(formMaxTokens.filter((_, i) => i !== index));
  }

  function addPrompt() {
    setFormPrompts([
      ...formPrompts,
      { id: `prompt_${Date.now()}`, label: `Prompt ${formPrompts.length + 1}`, content: "" },
    ]);
  }

  function updatePrompt(index: number, updates: Partial<SweepPromptVariant>) {
    const newPrompts = [...formPrompts];
    newPrompts[index] = { ...newPrompts[index], ...updates };
    setFormPrompts(newPrompts);
  }

  function removePrompt(index: number) {
    setFormPrompts(formPrompts.filter((_, i) => i !== index));
  }

  async function handleCreate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!formName.trim() || !formTaskID.trim()) return;

    setIsSubmitting(true);
    setError(null);
    try {
      const params: SweepParameters = {
        models: formModels,
        temperatures: formTemperatures,
        max_tokens: formMaxTokens,
        prompts: formPrompts,
      };
      const result = await createSweep({
        name: formName.trim(),
        description: formDescription.trim(),
        task_id: formTaskID.trim(),
        parameters: params,
      });
      toast.success(`${tc("success")}: ${t("createdWithRuns", { count: result.total_runs })}`);
      resetForm();
      await loadSweeps();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("createFailed"));
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleDelete(id: string) {
    if (!confirm(t("confirmDelete"))) return;
    try {
      await deleteSweep(id);
      toast.success(tc("success") + ": " + t("deleted"));
      await loadSweeps();
      if (expandedId === id) {
        setExpandedId(null);
        setSelectedSweep(null);
        setSweepMatrix(null);
        setMatrixSummary(null);
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("deleteFailed"));
    }
  }

  function getStatusBadge(status: string) {
    switch (status) {
      case "completed":
        return <span className="badge badge-completed"><CheckCircle2 size={12} /> {t("completed")}</span>;
      case "running":
      case "in_progress":
        return <span className="badge" style={{ background: "var(--warning)", color: "white" }}><Loader2 size={12} className="spin" /> {t("running")}</span>;
      case "failed":
        return <span className="badge badge-error"><AlertCircle size={12} /> {t("failed")}</span>;
      default:
        return <span className="badge badge-muted"><Pause size={12} /> {t("pending")}</span>;
    }
  }

  if (isLoading) {
    return (
      <div className="admin-content">
        <Breadcrumb />
        <div className="loading-pulse loading-pulse-long" style={{ marginBottom: 16 }} />
        <div className="skeleton-grid">
          <div className="skeleton-card" />
          <div className="skeleton-card" />
        </div>
      </div>
    );
  }

  const totalRuns = formModels.length * formTemperatures.length * formMaxTokens.length * formPrompts.length;

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
            {t("newSweep")}
          </button>
        </div>
      </div>

      {error && (
        <div className="form-error" style={{ marginBottom: 16 }}>
          {error}
        </div>
      )}

      {showCreateForm && (
        <div className="form-panel">
          <div className="form-header" style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
            <div>
              <p className="section-kicker">{t("title")}</p>
              <h2 className="form-title">{t("createSweep")}</h2>
            </div>
            <button
              className="btn btn-ghost btn-sm"
              onClick={resetForm}
              aria-label={tc("close")}
            >
              <X size={18} aria-hidden="true" />
            </button>
          </div>
          <form className="form-fields" onSubmit={(e) => void handleCreate(e)}>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 }}>
              <div>
                <label className="form-label" htmlFor="sweep-name">{t("sweepName")}</label>
                <input
                  id="sweep-name"
                  className="form-input"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  placeholder={t("sweepNamePlaceholder")}
                  required
                  autoFocus
                />
              </div>
              <div>
                <label className="form-label" htmlFor="sweep-task">{t("taskID")}</label>
                <input
                  id="sweep-task"
                  className="form-input"
                  value={formTaskID}
                  onChange={(e) => setFormTaskID(e.target.value)}
                  placeholder={t("taskIDPlaceholder")}
                  required
                />
              </div>
            </div>
            <div>
              <label className="form-label" htmlFor="sweep-desc">{tc("description")}</label>
              <input
                id="sweep-desc"
                className="form-input"
                value={formDescription}
                onChange={(e) => setFormDescription(e.target.value)}
                placeholder={t("sweepDescPlaceholder")}
              />
            </div>

            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr 1fr", gap: 16 }}>
              <div>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
                  <label className="form-label" style={{ marginBottom: 0 }}>{t("models")}</label>
                  <button type="button" className="btn btn-sm btn-ghost" onClick={addModel}>
                    <Plus size={12} />
                  </button>
                </div>
                <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
                  {formModels.map((model, i) => (
                    <div key={i} style={{ display: "flex", gap: 4 }}>
                      <select
                        className="form-select"
                        value={model}
                        onChange={(e) => updateModel(i, e.target.value)}
                        style={{ flex: 1 }}
                      >
                        {DEFAULT_MODELS.map((m) => (
                          <option key={m} value={m}>{m}</option>
                        ))}
                      </select>
                      {formModels.length > 1 && (
                        <button type="button" className="btn btn-ghost btn-sm" onClick={() => removeModel(i)}>
                          <X size={12} />
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              </div>

              <div>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
                  <label className="form-label" style={{ marginBottom: 0 }}>{t("temperature")}</label>
                  <button type="button" className="btn btn-sm btn-ghost" onClick={addTemperature}>
                    <Plus size={12} />
                  </button>
                </div>
                <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
                  {formTemperatures.map((temp, i) => (
                    <div key={i} style={{ display: "flex", gap: 4 }}>
                      <input
                        type="number"
                        className="form-input"
                        value={temp}
                        onChange={(e) => updateTemperature(i, parseFloat(e.target.value) || 0)}
                        step="0.1"
                        min="0"
                        max="2"
                        style={{ flex: 1 }}
                      />
                      {formTemperatures.length > 1 && (
                        <button type="button" className="btn btn-ghost btn-sm" onClick={() => removeTemperature(i)}>
                          <X size={12} />
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              </div>

              <div>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
                  <label className="form-label" style={{ marginBottom: 0 }}>{t("maxTokens")}</label>
                  <button type="button" className="btn btn-sm btn-ghost" onClick={addMaxTokens}>
                    <Plus size={12} />
                  </button>
                </div>
                <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
                  {formMaxTokens.map((tokens, i) => (
                    <div key={i} style={{ display: "flex", gap: 4 }}>
                      <input
                        type="number"
                        className="form-input"
                        value={tokens}
                        onChange={(e) => updateMaxTokens(i, parseInt(e.target.value) || 0)}
                        step="64"
                        min="1"
                        style={{ flex: 1 }}
                      />
                      {formMaxTokens.length > 1 && (
                        <button type="button" className="btn btn-ghost btn-sm" onClick={() => removeMaxTokens(i)}>
                          <X size={12} />
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              </div>

              <div>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
                  <label className="form-label" style={{ marginBottom: 0 }}>{t("prompts")}</label>
                  <button type="button" className="btn btn-sm btn-ghost" onClick={addPrompt}>
                    <Plus size={12} />
                  </button>
                </div>
                <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
                  {formPrompts.map((prompt, i) => (
                    <div key={i} style={{ display: "flex", gap: 4, alignItems: "center" }}>
                      <input
                        className="form-input"
                        value={prompt.label}
                        onChange={(e) => updatePrompt(i, { label: e.target.value })}
                        placeholder={t("promptLabel")}
                        style={{ flex: 1 }}
                      />
                      {formPrompts.length > 1 && (
                        <button type="button" className="btn btn-ghost btn-sm" onClick={() => removePrompt(i)}>
                          <X size={12} />
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div style={{ padding: 12, background: "var(--surface)", borderRadius: 4, fontSize: "var(--text-sm)" }}>
              <strong>{t("totalRuns")}:</strong> {totalRuns} ({formModels.length} × {formTemperatures.length} × {formMaxTokens.length} × {formPrompts.length})
            </div>

            <div className="form-actions">
              <button type="button" className="btn btn-secondary" onClick={resetForm}>
                {tc("cancel")}
              </button>
              <button type="submit" className="btn btn-primary" disabled={isSubmitting}>
                {isSubmitting ? t("creating") : <><Play size={14} /> {t("create")}</>}
              </button>
            </div>
          </form>
        </div>
      )}

      {sweeps.length === 0 ? (
        <div className="empty-state">
          <div className="empty-state-icon">
            <Crosshair size={48} aria-hidden="true" />
          </div>
          <h3 className="empty-state-title">{t("noSweeps")}</h3>
          <p className="empty-state-description">
            {t("noSweepsDescription")}
          </p>
          <div className="empty-state-action" style={{ marginTop: 16 }}>
            <button className="btn btn-primary" onClick={() => setShowCreateForm(true)}>
              <Plus size={16} aria-hidden="true" />
              {t("newSweep")}
            </button>
          </div>
        </div>
      ) : (
        <div className="panel">
          <div className="gateway-table-wrapper">
            <table className="gateway-table">
              <thead>
                <tr>
                  <th style={{ width: 40 }}></th>
                  <th>{t("name")}</th>
                  <th>{t("taskID")}</th>
                  <th>{t("progress")}</th>
                  <th>{t("status")}</th>
                  <th style={{ width: 100 }}></th>
                </tr>
              </thead>
              <tbody>
                {sweeps.map((sweep) => {
                  const isExpanded = expandedId === sweep.id;
                  const detail = isExpanded ? selectedSweep : null;
                  const matrix = isExpanded ? sweepMatrix : null;
                  const summary = isExpanded ? matrixSummary : null;

                  return (
                    <>
                      <tr key={sweep.id}>
                        <td>
                          <button
                            className="btn btn-ghost btn-sm"
                            onClick={() => toggleExpanded(sweep.id)}
                          >
                            {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                          </button>
                        </td>
                        <td>
                          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                            <Crosshair size={16} aria-hidden="true" style={{ color: "var(--muted)", flexShrink: 0 }} />
                            <div>
                              <div style={{ fontWeight: 600, color: "var(--ink-strong)", fontSize: "var(--text-sm)" }}>
                                {sweep.name}
                              </div>
                              {sweep.description && (
                                <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)", marginTop: 2 }}>
                                  {sweep.description}
                                </div>
                              )}
                            </div>
                          </div>
                        </td>
                        <td>
                          <code style={{ fontSize: "var(--text-xs)" }}>{sweep.task_id}</code>
                        </td>
                        <td>
                          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                            <div style={{ flex: 1, height: 4, background: "var(--surface)", borderRadius: 2 }}>
                              <div
                                style={{
                                  width: `${sweep.total_runs > 0 ? (sweep.completed / sweep.total_runs) * 100 : 0}%`,
                                  height: "100%",
                                  background: sweep.status === "completed" ? "var(--success)" : "var(--primary)",
                                  borderRadius: 2,
                                }}
                              />
                            </div>
                            <span style={{ fontSize: "var(--text-xs)", color: "var(--muted)", minWidth: 50 }}>
                              {sweep.completed}/{sweep.total_runs}
                            </span>
                          </div>
                        </td>
                        <td>
                          {getStatusBadge(sweep.status)}
                        </td>
                        <td>
                          <button
                            className="btn btn-ghost btn-sm"
                            onClick={() => handleDelete(sweep.id)}
                            title={t("delete")}
                          >
                            <Trash2 size={14} aria-hidden="true" />
                          </button>
                        </td>
                      </tr>
                      {isExpanded && detail && (
                        <tr key={`${sweep.id}-detail`}>
                          <td colSpan={6} style={{ padding: 16, background: "var(--background-subtle)" }}>
                            {summary && (
                              <div style={{ display: "grid", gridTemplateColumns: "repeat(6, 1fr)", gap: 12, marginBottom: 16 }}>
                                <div style={{ padding: 8, background: "var(--surface)", borderRadius: 4, textAlign: "center" }}>
                                  <div style={{ fontSize: "var(--text-lg)", fontWeight: 600 }}>{summary.total}</div>
                                  <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>{t("total")}</div>
                                </div>
                                <div style={{ padding: 8, background: "var(--surface)", borderRadius: 4, textAlign: "center" }}>
                                  <div style={{ fontSize: "var(--text-lg)", fontWeight: 600 }}>{summary.completed}</div>
                                  <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>{t("completed")}</div>
                                </div>
                                <div style={{ padding: 8, background: "var(--surface)", borderRadius: 4, textAlign: "center" }}>
                                  <div style={{ fontSize: "var(--text-lg)", fontWeight: 600 }}>{summary.pending}</div>
                                  <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>{t("pending")}</div>
                                </div>
                                <div style={{ padding: 8, background: "var(--surface)", borderRadius: 4, textAlign: "center" }}>
                                  <div style={{ fontSize: "var(--text-lg)", fontWeight: 600 }}>{summary.avg_score.toFixed(2)}</div>
                                  <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>{t("avgScore")}</div>
                                </div>
                                <div style={{ padding: 8, background: "var(--surface)", borderRadius: 4, textAlign: "center" }}>
                                  <div style={{ fontSize: "var(--text-lg)", fontWeight: 600 }}>{summary.max_score.toFixed(2)}</div>
                                  <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>{t("maxScore")}</div>
                                </div>
                                <div style={{ padding: 8, background: "var(--surface)", borderRadius: 4, textAlign: "center" }}>
                                  <div style={{ fontSize: "var(--text-lg)", fontWeight: 600 }}>{summary.min_score.toFixed(2)}</div>
                                  <div style={{ fontSize: "var(--text-xs)", color: "var(--muted)" }}>{t("minScore")}</div>
                                </div>
                              </div>
                            )}
                            {matrix && matrix.length > 0 && (
                              <div style={{ overflowX: "auto" }}>
                                <table className="gateway-table" style={{ fontSize: "var(--text-xs)" }}>
                                  <thead>
                                    <tr>
                                      <th>{t("model")}</th>
                                      <th>{t("temperature")}</th>
                                      <th>{t("maxTokens")}</th>
                                      <th>{t("prompt")}</th>
                                      <th>{t("status")}</th>
                                      <th>{t("score")}</th>
                                    </tr>
                                  </thead>
                                  <tbody>
                                    {matrix.flat().map((cell, i) => (
                                      <tr key={i}>
                                        <td><code>{cell.model}</code></td>
                                        <td>{cell.temperature}</td>
                                        <td>{cell.max_tokens}</td>
                                        <td>{cell.prompt_label}</td>
                                        <td>{getStatusBadge(cell.status)}</td>
                                        <td>{cell.score !== undefined ? cell.score.toFixed(2) : "-"}</td>
                                      </tr>
                                    ))}
                                  </tbody>
                                </table>
                              </div>
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
    </div>
  );
}