"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { createSkill } from "../lib/api";

export default function CreateSkillForm() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [ownerTeam, setOwnerTeam] = useState("");
  const [riskLevel, setRiskLevel] = useState("medium");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await createSkill({ name, owner_team: ownerTeam, risk_level: riskLevel });
      setName("");
      setOwnerTeam("");
      setRiskLevel("medium");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create skill");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">Registry</p>
        <h2>Register new skill</h2>
      </div>

      {error && <p className="form-error">{error}</p>}

      <div className="form-fields">
        <label className="form-label">
          Skill name
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. restart_service_skill"
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          Owner team
          <input
            type="text"
            value={ownerTeam}
            onChange={(e) => setOwnerTeam(e.target.value)}
            placeholder="e.g. platform_team"
            required
            className="form-input"
          />
        </label>

        <label className="form-label">
          Risk level
          <select
            value={riskLevel}
            onChange={(e) => setRiskLevel(e.target.value)}
            className="form-select"
          >
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
          </select>
        </label>
      </div>

      <button type="submit" disabled={loading} className="form-submit">
        {loading ? "Registering..." : "Register skill"}
      </button>
    </form>
  );
}
