"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { runTaskEvaluation, Task } from "../lib/api";
import { useTranslations } from "next-intl";

type Props = {
  tasks: Task[];
  onCreated?: () => void;
};

export default function RunEvaluationForm({ tasks, onCreated }: Props) {
  const t = useTranslations("evaluations");
  const te = useTranslations("errors");
  const router = useRouter();
  const [taskId, setTaskId] = useState("");
  const [agentId, setAgentId] = useState("");
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    let parsedInput: Record<string, unknown> = {};
    if (input.trim()) {
      try {
        parsedInput = JSON.parse(input);
      } catch {
        setError(te("inputInvalid"));
        setLoading(false);
        return;
      }
    }

    try {
      await runTaskEvaluation({
        task_id: taskId,
        agent_id: agentId,
        input: parsedInput,
      });
      setTaskId("");
      setAgentId("");
      setInput("");
      onCreated?.();
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to run evaluation");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("runEvaluation")}</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <div className="form-fields">
        <label className="form-label">
          {t("taskOverview").split("(")[0].trim()}
          <select
            value={taskId}
            onChange={(e) => setTaskId(e.target.value)}
            required
            className="form-select"
          >
            <option value="">{t("select").split("...")[0]}...</option>
            {tasks.map((t) => (
              <option key={t.id} value={t.id}>
                {t.name}
              </option>
            ))}
          </select>
        </label>

        <label className="form-label">
          {t("agentId")}
          <input
            type="text"
            value={agentId}
            onChange={(e) => setAgentId(e.target.value)}
            placeholder={t("agentIdPlaceholder")}
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("input").split("(")[0].trim()}
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder='{"query": "test input"}'
            rows={3}
            className="form-textarea"
          />
        </label>
      </div>

      <button type="submit" disabled={loading} className="form-submit">
        {loading ? t("running") : t("run")}
      </button>
    </form>
  );
}