"use client";

import { useState, FormEvent } from "react";
import { updateTask, Skill, Task, CreateTaskRequest } from "../lib/api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";
import { useToast } from "../../components/ToastProvider";

type Props = {
  task: Task;
  skills: Skill[];
  onCancel: () => void;
  onUpdated: () => void;
};

type UpdateTaskVariables = {
  id: string;
} & Partial<CreateTaskRequest>;

async function updateTaskWrapper(vars: UpdateTaskVariables) {
  const { id, ...req } = vars;
  return updateTask(id, req);
}

export default function EditTaskForm({ task, skills, onCancel, onUpdated }: Props) {
  const t = useTranslations("tasks");
  const te = useTranslations("errors");
  const toast = useToast();
  const [name, setName] = useState(task.name);
  const [description, setDescription] = useState(task.description);
  const [skillId, setSkillId] = useState(task.skill_id);
  const [difficulty, setDifficulty] = useState<string>(task.difficulty);
  const [tags, setTags] = useState(task.tags.join(", "));
  const [testCases, setTestCases] = useState(JSON.stringify(task.test_cases, null, 2));

  const updateMutation = useApiMutation(updateTaskWrapper, {
    successMessage: "Task updated",
    queryKeysToInvalidate: [["tasks"]],
    onSuccess: () => {
      onUpdated();
    },
  });

  function handleSubmit(e: FormEvent) {
    e.preventDefault();

    let parsedTestCases: Array<{ input: Record<string, unknown>; expected: unknown }> = [];
    if (testCases.trim()) {
      try {
        parsedTestCases = JSON.parse(testCases);
      } catch {
        toast.error(te("testCasesInvalid"));
        return;
      }
    }

    updateMutation.mutate({
      id: task.id,
      name,
      description,
      skill_id: skillId,
      tags: tags.split(",").map((t) => t.trim()).filter(Boolean),
      difficulty,
      test_cases: parsedTestCases,
    });
  }

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">{t("eyebrow")}</p>
        <h2>{t("editTask", { name: task.name })}</h2>
      </div>

      {updateMutation.isError && <p className="form-error">{updateMutation.error?.message}</p>}

      <div className="form-fields">
        <label className="form-label">
          {t("taskName")}
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("description")}
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={2}
            className="form-textarea"
          />
        </label>

        <label className="form-label">
          {t("linkedSkill")}
          <select
            value={skillId}
            onChange={(e) => setSkillId(e.target.value)}
            className="form-select"
          >
            <option value="">{t("noSkill")}</option>
            {skills.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
        </label>

        <label className="form-label">
          {t("difficulty")}
          <select
            value={difficulty}
            onChange={(e) => setDifficulty(e.target.value)}
            className="form-select"
          >
            <option value="easy">{t("easy")}</option>
            <option value="medium">{t("medium")}</option>
            <option value="hard">{t("hard")}</option>
          </select>
        </label>

        <label className="form-label">
          {t("tags")}
          <input
            type="text"
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("testCases")}
          <textarea
            value={testCases}
            onChange={(e) => setTestCases(e.target.value)}
            rows={3}
            className="form-textarea"
          />
        </label>
      </div>

      <div className="form-actions">
        <button type="submit" disabled={updateMutation.isPending} className="form-submit" style={{ flex: 1 }}>
          {updateMutation.isPending ? t("saving") : t("saveChanges")}
        </button>
        <button type="button" onClick={onCancel} className="btn btn-secondary" style={{ flex: 1 }}>
          {t("cancel")}
        </button>
      </div>
    </form>
  );
}