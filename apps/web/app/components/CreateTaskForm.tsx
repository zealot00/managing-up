"use client";

import { useState, FormEvent } from "react";
import { createTask, Skill } from "../lib/api";
import { useTranslations } from "next-intl";
import { useApiMutation } from "../lib/use-mutations";
import { useToast } from "../../components/ToastProvider";

type Props = {
  skills: Skill[];
  onCreated?: () => void;
};

export default function CreateTaskForm({ skills, onCreated }: Props) {
  const t = useTranslations("tasks");
  const te = useTranslations("errors");
  const toast = useToast();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [skillId, setSkillId] = useState("");
  const [difficulty, setDifficulty] = useState("medium");
  const [tags, setTags] = useState("");
  const [testCases, setTestCases] = useState("");

  const createMutation = useApiMutation(createTask, {
    successMessage: "Task created",
    queryKeysToInvalidate: [["tasks"]],
    onSuccess: () => {
      setName("");
      setDescription("");
      setSkillId("");
      setDifficulty("medium");
      setTags("");
      setTestCases("");
      onCreated?.();
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

    createMutation.mutate({
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
        <h2>{t("createTask")}</h2>
      </div>

      {createMutation.isError && <p className="form-error">{createMutation.error?.message}</p>}

      <div className="form-fields">
        <label className="form-label">
          {t("taskName")}
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t("taskNamePlaceholder")}
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("description")}
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="What does this task evaluate?"
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
            placeholder={t("tagsPlaceholder")}
            className="form-input"
          />
        </label>

        <label className="form-label">
          {t("testCases")}
          <textarea
            value={testCases}
            onChange={(e) => setTestCases(e.target.value)}
            placeholder={t("testCasesPlaceholder")}
            rows={3}
            className="form-textarea"
          />
        </label>
      </div>

      <button type="submit" disabled={createMutation.isPending} className="form-submit">
        {createMutation.isPending ? t("creating") : t("create")}
      </button>
    </form>
  );
}