"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { createExecution, Skill } from "../lib/api";

type Props = {
  skills: Skill[];
};

export default function TriggerExecutionForm({ skills }: Props) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [skillId, setSkillId] = useState("");
  const [triggeredBy, setTriggeredBy] = useState("");
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
        setError("Input must be valid JSON");
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
      setIsOpen(false);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to trigger execution");
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="trigger-btn"
      >
        {isOpen ? "Cancel" : "Trigger Execution"}
      </button>

      {isOpen && (
        <form onSubmit={handleSubmit} className="form-panel">
          <div className="panel-header">
            <p className="section-kicker">Execution Timeline</p>
            <h2>Trigger execution</h2>
          </div>

          {error && <p className="form-error">{error}</p>}

          <div className="form-fields">
            <label className="form-label">
              Skill
              <select
                value={skillId}
                onChange={(e) => setSkillId(e.target.value)}
                required
                className="form-select"
              >
                <option value="">Select a skill...</option>
                {skills.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name} ({s.owner_team})
                  </option>
                ))}
              </select>
            </label>

            <label className="form-label">
              Triggered by
              <input
                type="text"
                value={triggeredBy}
                onChange={(e) => setTriggeredBy(e.target.value)}
                placeholder="e.g. platform_operator"
                required
                className="form-input"
              />
            </label>

            <label className="form-label">
              Input (JSON)
              <textarea
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder='{"server_id": "srv-001"}'
                rows={3}
                className="form-textarea"
              />
            </label>
          </div>

          <button type="submit" disabled={loading} className="form-submit">
            {loading ? "Triggering..." : "Trigger execution"}
          </button>
        </form>
      )}
    </>
  );
}
