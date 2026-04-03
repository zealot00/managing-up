"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { runExperiment, checkRegression, Experiment } from "../lib/api";
import { useTranslations } from "next-intl";

type Props = {
  exp: Experiment;
  onRun?: () => void;
};

export default function ExperimentCardWithActions({ exp, onRun }: Props) {
  const t = useTranslations("experiments");
  const tc = useTranslations("common");
  const router = useRouter();
  const [running, setRunning] = useState(false);
  const [error, setError] = useState("");

  async function handleRun() {
    setRunning(true);
    setError("");
    try {
      await runExperiment(exp.id);
      onRun?.();
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to run experiment");
    } finally {
      setRunning(false);
    }
  }

  return (
    <article className="eval-card">
      <div className="eval-card-header">
        <div>
          <h3 className="eval-card-title">{exp.name}</h3>
          <p className="eval-card-meta">{exp.description || tc("noData")}</p>
        </div>
        <span className={`badge badge-${exp.status === "completed" ? "succeeded" : exp.status === "running" ? "running" : "muted"}`}>
          {exp.status}
        </span>
      </div>
      <div className="tags">
        <span className="tag">{t("tasksCount", { count: exp.task_ids.length })}</span>
        <span className="tag">{t("agentsCount", { count: exp.agent_ids.length })}</span>
      </div>
      {error && <p className="form-error" style={{ marginTop: "var(--space-2)" }}>{error}</p>}
      <div className="eval-card-footer">
        <span>{tc("createdAt")}: {new Date(exp.created_at).toLocaleString()}</span>
        {exp.status === "pending" && (
          <button
            className="btn btn-sm btn-primary"
            onClick={handleRun}
            disabled={running}
          >
            {running ? t("running") : t("run")}
          </button>
        )}
      </div>
    </article>
  );
}