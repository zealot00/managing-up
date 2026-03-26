"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { createSkillVersion, Skill } from "../lib/api";

type Props = {
  skills: Skill[];
};

export default function CreateSkillVersionForm({ skills }: Props) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [skillId, setSkillId] = useState("");
  const [version, setVersion] = useState("");
  const [changeSummary, setChangeSummary] = useState("");
  const [approvalRequired, setApprovalRequired] = useState(false);
  const [specYaml, setSpecYaml] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await createSkillVersion({
        skill_id: skillId,
        version,
        change_summary: changeSummary,
        approval_required: approvalRequired,
        spec_yaml: specYaml,
      });
      setSkillId("");
      setVersion("");
      setChangeSummary("");
      setApprovalRequired(false);
      setSpecYaml("");
      setIsOpen(false);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create skill version");
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <button onClick={() => setIsOpen(!isOpen)} className="trigger-btn">
        {isOpen ? "Cancel" : "Create New Version"}
      </button>

      {isOpen && (
        <form onSubmit={handleSubmit} className="form-panel">
          <div className="panel-header">
            <p className="section-kicker">Version Control</p>
            <h2>Create skill version</h2>
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
              Version
              <input
                type="text"
                value={version}
                onChange={(e) => setVersion(e.target.value)}
                placeholder="e.g. 1.0.0"
                required
                className="form-input"
              />
            </label>

            <label className="form-label">
              Change summary
              <input
                type="text"
                value={changeSummary}
                onChange={(e) => setChangeSummary(e.target.value)}
                placeholder="e.g. Added validation step for input sanitization"
                required
                className="form-input"
              />
            </label>

            <label className="form-label" style={{ display: "flex", alignItems: "center", gap: 12 }}>
              <input
                type="checkbox"
                checked={approvalRequired}
                onChange={(e) => setApprovalRequired(e.target.checked)}
                style={{ width: 18, height: 18 }}
              />
              Require approval before execution
            </label>

            <label className="form-label">
              Spec YAML
              <textarea
                value={specYaml}
                onChange={(e) => setSpecYaml(e.target.value)}
                placeholder={`name: my_skill
description: Does something useful
steps:\n  - id: step1\n    name: First step\n    tool: execute_command\n    args:\n      command: echo "hello"`}
                rows={10}
                required
                className="form-textarea"
                style={{ fontFamily: "monospace", fontSize: "0.85rem" }}
              />
            </label>
          </div>

          <button type="submit" disabled={loading} className="form-submit">
            {loading ? "Creating..." : "Create version"}
          </button>
        </form>
      )}
    </>
  );
}
