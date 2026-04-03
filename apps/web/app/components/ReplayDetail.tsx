"use client";

import { ReplaySnapshot } from "../lib/api";

type Props = {
  snapshot: ReplaySnapshot;
  onClose: () => void;
};

export default function ReplayDetail({ snapshot, onClose }: Props) {
  return (
    <div className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">Replay Layer</p>
        <h2>Snapshot Detail</h2>
      </div>

      <div className="detail-grid">
        <div className="detail-row">
          <span className="detail-label">ID</span>
          <span className="detail-value">{snapshot.id}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">Execution ID</span>
          <span className="detail-value">{snapshot.execution_id}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">Skill ID</span>
          <span className="detail-value">{snapshot.skill_id}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">Version</span>
          <span className="detail-value">{snapshot.skill_version}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">Step Index</span>
          <span className="detail-value">{snapshot.step_index}</span>
        </div>
        <div className="detail-row">
          <span className="detail-label">Deterministic Seed</span>
          <span className="detail-value" style={{ fontFamily: "monospace" }}>
            {snapshot.deterministic_seed}
          </span>
        </div>
      </div>

      <div className="panel-header" style={{ marginTop: "var(--space-6)" }}>
        <p className="section-kicker">State</p>
        <h2>State Snapshot</h2>
      </div>
      <pre className="json-block">
        {JSON.stringify(snapshot.state_snapshot, null, 2)}
      </pre>

      <div className="panel-header" style={{ marginTop: "var(--space-6)" }}>
        <p className="section-kicker">Input</p>
        <h2>Input Seed</h2>
      </div>
      <pre className="json-block">
        {JSON.stringify(snapshot.input_seed, null, 2)}
      </pre>

      <div className="form-actions" style={{ marginTop: "var(--space-6)" }}>
        <button onClick={onClose} className="btn btn-secondary" style={{ flex: 1 }}>
          Close
        </button>
      </div>
    </div>
  );
}
