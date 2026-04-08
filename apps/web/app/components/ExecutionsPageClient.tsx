"use client";

import { useState, useMemo } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Execution, Skill, createExecution } from "../lib/api";
import { useTranslations } from "next-intl";
import { useToast } from "../../components/ToastProvider";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";

type Props = {
  executions: { items: Execution[] };
  skills: Skill[];
};

export default function ExecutionsPageClient({ executions, skills }: Props) {
  const t = useTranslations("executions");
  const tc = useTranslations("common");
  const router = useRouter();
  const toast = useToast();
  const [showModal, setShowModal] = useState(false);
  const [skillId, setSkillId] = useState("");
  const [triggeredBy, setTriggeredBy] = useState("");
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState("");

  const filteredExecutions = useMemo(() => {
    return executions.items.filter((exec) => {
      const matchesStatus = statusFilter === "all" || exec.status === statusFilter;
      const matchesSearch = searchQuery === "" ||
        exec.skill_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        exec.triggered_by.toLowerCase().includes(searchQuery.toLowerCase());
      return matchesStatus && matchesSearch;
    });
  }, [executions.items, statusFilter, searchQuery]);

  const uniqueStatuses = useMemo(() => {
    const statuses = new Set(executions.items.map((e) => e.status));
    return Array.from(statuses).sort();
  }, [executions.items]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    let parsedInput: Record<string, unknown> = {};
    if (input.trim()) {
      try {
        parsedInput = JSON.parse(input);
      } catch {
        setError(tc("errors.inputInvalid"));
        setLoading(false);
        return;
      }
    }

    try {
      await createExecution({
        skill_id: skillId,
        triggered_by: triggeredBy,
        input: parsedInput,
      });
      setSkillId("");
      setTriggeredBy("");
      setInput("");
      setShowModal(false);
      toast.success(tc("success") + ": Execution triggered");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to trigger execution");
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={t("lede")}
        actions={
          <button
            className="btn btn-primary"
            onClick={() => setShowModal(true)}
          >
            + {t("triggerExecution")}
          </button>
        }
      />

      <div style={{ display: "flex", gap: "var(--space-4)", marginBottom: "var(--space-6)", flexWrap: "wrap", alignItems: "center" }}>
        <div style={{ flex: "1 1 200px", maxWidth: 280 }}>
          <input
            type="text"
            placeholder="Search by skill or operator..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="form-input"
            style={{ width: "100%" }}
          />
        </div>

        <div style={{ display: "flex", gap: "var(--space-2)", flexWrap: "wrap" }}>
          <button
            onClick={() => setStatusFilter("all")}
            className={`btn btn-sm ${statusFilter === "all" ? "btn-primary" : "btn-secondary"}`}
          >
            All
          </button>
          {uniqueStatuses.map((status) => (
            <button
              key={status}
              onClick={() => setStatusFilter(status)}
              className={`btn btn-sm ${statusFilter === status ? "btn-primary" : "btn-secondary"}`}
            >
              {status}
            </button>
          ))}
        </div>

        {(searchQuery || statusFilter !== "all") && (
          <button
            onClick={() => { setSearchQuery(""); setStatusFilter("all"); }}
            className="btn btn-ghost btn-sm"
          >
            Clear
          </button>
        )}
      </div>

      <section className="panel">
        <div className="panel-header">
          <p className="section-kicker">{t("runs")}</p>
          <h2 className="panel-title">
            {filteredExecutions.length === executions.items.length
              ? t("executionQueue")
              : `${filteredExecutions.length} of ${executions.items.length} runs`}
          </h2>
        </div>
        <div className="list">
          {filteredExecutions.length === 0 ? (
            <EmptyState title={t("noExecutions")} />
          ) : (
            filteredExecutions.map((execution) => (
              <article className="list-card" key={execution.id}>
                <div className="list-card-main">
                  <h3 className="list-card-title">{execution.skill_name}</h3>
                  <p className="list-card-meta">
                    {execution.current_step_id} · {t("triggeredBy")} {execution.triggered_by}
                  </p>
                </div>
                <div className="list-card-actions">
                  <Link href={`/executions/${execution.id}`} className="trace-link">
                    {t("viewTrace")}
                  </Link>
                  <span className={`badge badge-${execution.status}`}>{execution.status}</span>
                </div>
              </article>
            ))
          )}
        </div>
      </section>

      {showModal && (
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
            if (e.target === e.currentTarget) setShowModal(false);
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
                  {t("trigger")}
                </h2>
              </div>
              <button
                onClick={() => setShowModal(false)}
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

            <form onSubmit={handleSubmit}>
              <div className="form-fields">
                <label className="form-label">
                  {t("skill")}
                  <select
                    value={skillId}
                    onChange={(e) => setSkillId(e.target.value)}
                    required
                    className="form-select"
                  >
                    <option value="">{t("selectSkill")}</option>
                    {skills.map((s) => (
                      <option key={s.id} value={s.id}>
                        {s.name} ({s.owner_team})
                      </option>
                    ))}
                  </select>
                </label>

                <label className="form-label">
                  {t("triggeredBy")}
                  <input
                    type="text"
                    value={triggeredBy}
                    onChange={(e) => setTriggeredBy(e.target.value)}
                    placeholder={t("triggeredByPlaceholder")}
                    required
                    className="form-input"
                  />
                </label>

                <label className="form-label">
                  {t("input")}
                  <textarea
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    placeholder={t("inputPlaceholder")}
                    rows={3}
                    className="form-textarea"
                  />
                </label>
              </div>

              <button type="submit" disabled={loading} className="form-submit" style={{ marginTop: "var(--space-4)" }}>
                {loading ? t("triggering") : t("trigger")}
              </button>
            </form>
          </div>
        </div>
      )}
    </>
  );
}
