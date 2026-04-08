"use client";

import { runExperiment, Experiment } from "../lib/api";
import { useTranslations } from "next-intl";
import { Badge } from "./ui/Badge";
import { useApiMutation } from "../lib/use-mutations";

type Props = {
  exp: Experiment;
  onRun?: () => void;
};

export default function ExperimentCardWithActions({ exp, onRun }: Props) {
  const t = useTranslations("experiments");
  const tc = useTranslations("common");

  const runExperimentMutation = useApiMutation(runExperiment, {
    queryKeysToInvalidate: [["experiments"]],
    onSuccess: () => {
      onRun?.();
    },
  });

  function handleRun() {
    runExperimentMutation.mutate(exp.id);
  }

  return (
    <article className="eval-card">
      <div className="eval-card-header">
        <div>
          <h3 className="eval-card-title">{exp.name}</h3>
          <p className="eval-card-meta">{exp.description || tc("noData")}</p>
        </div>
        <Badge variant={exp.status as "running" | "completed" | "pending" | "failed" | "muted"}>
          {exp.status}
        </Badge>
      </div>
      <div className="tags">
        <span className="tag">{t("tasksCount", { count: exp.task_ids.length })}</span>
        <span className="tag">{t("agentsCount", { count: exp.agent_ids.length })}</span>
      </div>
      {runExperimentMutation.error && <p className="form-error" style={{ marginTop: "var(--space-2)" }}>{runExperimentMutation.error.message}</p>}
      <div className="eval-card-footer">
        <span>{tc("createdAt")}: {new Date(exp.created_at).toLocaleString()}</span>
        {exp.status === "pending" && (
          <button
            className="btn btn-sm btn-primary"
            onClick={handleRun}
            disabled={runExperimentMutation.isPending}
          >
            {runExperimentMutation.isPending ? t("running") : t("run")}
          </button>
        )}
      </div>
    </article>
  );
}