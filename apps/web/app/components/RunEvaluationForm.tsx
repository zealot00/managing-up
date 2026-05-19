"use client";

import { useState, FormEvent } from "react";
import { runTaskEvaluation, Task } from "../lib/api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";

type Props = {
  tasks: Task[];
  onCreated?: () => void;
};

export default function RunEvaluationForm({ tasks, onCreated }: Props) {
  const t = useTranslations("evaluations");
  const tc = useTranslations("common");
  const te = useTranslations("errors");
  const [taskId, setTaskId] = useState("");
  const [agentId, setAgentId] = useState("");
  const [input, setInput] = useState("");
  const [localError, setLocalError] = useState("");

  const runEvaluationMutation = useApiMutation(runTaskEvaluation, {
    queryKeysToInvalidate: [["task-executions"]],
    successMessage: tc("success") + ": Evaluation started",
    onSuccess: () => {
      setTaskId("");
      setAgentId("");
      setInput("");
      onCreated?.();
    },
  });

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLocalError("");

    let parsedInput: Record<string, unknown> = {};
    if (input.trim()) {
      try {
        parsedInput = JSON.parse(input);
      } catch {
        setLocalError(te("inputInvalid"));
        return;
      }
    }

    runEvaluationMutation.mutate({
      task_id: taskId,
      agent_id: agentId,
      input: parsedInput,
    });
  }

  const error = runEvaluationMutation.error?.message || localError;

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("runEvaluation")}</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <div className="form-fields">
        <label className="form-label" htmlFor="task_id">
          {t("taskOverviewShort")}
          <select
            id="task_id"
            value={taskId}
            onChange={(e) => setTaskId(e.target.value)}
            required
            className="form-select"
          >
            <option value="">{tc("select")}...</option>
            {tasks.map((t) => (
              <option key={t.id} value={t.id}>
                {t.name}
              </option>
            ))}
          </select>
        </label>

        <label className="form-label" htmlFor="agent_id">
          {t("agentId")}
          <input
            type="text"
            id="agent_id"
            value={agentId}
            onChange={(e) => setAgentId(e.target.value)}
            placeholder={t("agentIdPlaceholder")}
            required
            className="form-input"
          />
        </label>

        <label className="form-label" htmlFor="eval_input">
          {t("input")}
          <textarea
            id="eval_input"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder={t("inputPlaceholder")}
            rows={2}
            className="form-textarea"
          />
        </label>
      </div>

      <button type="submit" disabled={runEvaluationMutation.isPending} className="form-submit">
        {runEvaluationMutation.isPending ? t("starting") : t("startEvaluation")}
      </button>
    </form>
  );
}