"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { updateTask, getSkills, Skill, Task } from "../lib/api";

type Props = {
  task: Task;
  skills: Skill[];
  onCancel: () => void;
  onUpdated: () => void;
};

export default function EditTaskForm({ task, skills, onCancel, onUpdated }: Props) {
  const router = useRouter();
  const [name, setName] = useState(task.name);
  const [description, setDescription] = useState(task.description);
  const [skillId, setSkillId] = useState(task.skill_id);
  const [difficulty, setDifficulty] = useState<string>(task.difficulty);
  const [tags, setTags] = useState(task.tags.join(", "));
  const [testCases, setTestCases] = useState(JSON.stringify(task.test_cases, null, 2));
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    let parsedTestCases: Array<{ input: Record<string, unknown>; expected: unknown }> = [];
    if (testCases.trim()) {
      try {
        parsedTestCases = JSON.parse(testCases);
      } catch {
        setError("Test cases must be valid JSON array");
        setLoading(false);
        return;
      }
    }

    try {
      await updateTask(task.id, {
        name,
        description,
        skill_id: skillId,
        tags: tags.split(",").map((t) => t.trim()).filter(Boolean),
        difficulty,
        test_cases: parsedTestCases,
      });
      onUpdated();
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update task");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">Task Registry</p>
        <h2>Edit task: {task.name}</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <div className="form-fields">
        <label className="form-label">
          Task name
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          Description
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={2}
            className="form-textarea"
          />
        </label>

        <label className="form-label">
          Linked skill
          <select
            value={skillId}
            onChange={(e) => setSkillId(e.target.value)}
            className="form-select"
          >
            <option value="">No skill</option>
            {skills.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
        </label>

        <label className="form-label">
          Difficulty
          <select
            value={difficulty}
            onChange={(e) => setDifficulty(e.target.value)}
            className="form-select"
          >
            <option value="easy">Easy</option>
            <option value="medium">Medium</option>
            <option value="hard">Hard</option>
          </select>
        </label>

        <label className="form-label">
          Tags (comma-separated)
          <input
            type="text"
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            className="form-input"
          />
        </label>

        <label className="form-label">
          Test cases (JSON array)
          <textarea
            value={testCases}
            onChange={(e) => setTestCases(e.target.value)}
            rows={3}
            className="form-textarea"
          />
        </label>
      </div>

      <div className="form-actions">
        <button type="submit" disabled={loading} className="form-submit" style={{ flex: 1 }}>
          {loading ? "Saving..." : "Save changes"}
        </button>
        <button type="button" onClick={onCancel} className="btn btn-secondary" style={{ flex: 1 }}>
          Cancel
        </button>
      </div>
    </form>
  );
}
