"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { createExperiment, Task } from "../lib/api";
import { useTranslations } from "next-intl";

type Props = {
  tasks: Task[];
  onCreated?: () => void;
};

export default function CreateExperimentForm({ tasks, onCreated }: Props) {
  const t = useTranslations("experiments");
  const tc = useTranslations("common");
  const router = useRouter();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [taskIds, setTaskIds] = useState("");
  const [agentIds, setAgentIds] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await createExperiment({
        name,
        description,
        task_ids: taskIds.split(",").map((t) => t.trim()).filter(Boolean),
        agent_ids: agentIds.split(",").map((a) => a.trim()).filter(Boolean),
      });
      setName("");
      setDescription("");
      setTaskIds("");
      setAgentIds("");
      onCreated?.();
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create experiment");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("createExperiment")}</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <div className="form-fields">
        <label className="form-label">
          {t("experimentName")}
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t("experimentNamePlaceholder")}
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          {tc("description")}
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="What are you testing?"
            rows={2}
            className="form-textarea"
          />
        </label>

        <label className="form-label">
          {t("taskIds")}
          <input
            type="text"
            value={taskIds}
            onChange={(e) => setTaskIds(e.target.value)}
            placeholder={t("taskIdsPlaceholder")}
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("agentIds")}
          <input
            type="text"
            value={agentIds}
            onChange={(e) => setAgentIds(e.target.value)}
            placeholder={t("agentIdsPlaceholder")}
            className="form-input"
          />
        </label>
      </div>

      <button type="submit" disabled={loading} className="form-submit">
        {loading ? t("creating") : t("createExperiment")}
      </button>
    </form>
  );
}