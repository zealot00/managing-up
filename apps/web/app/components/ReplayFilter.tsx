"use client";

import { useState } from "react";
import { ReplaySnapshot } from "../lib/api";

type Props = {
  onFilter: (filtered: ReplaySnapshot[]) => void;
  allSnapshots: ReplaySnapshot[];
};

export default function ReplayFilter({ onFilter, allSnapshots }: Props) {
  const [executionId, setExecutionId] = useState("");

  function handleFilter() {
    if (!executionId.trim()) {
      onFilter(allSnapshots);
      return;
    }
    const filtered = allSnapshots.filter((s) =>
      s.execution_id.toLowerCase().includes(executionId.toLowerCase())
    );
    onFilter(filtered);
  }

  return (
    <div className="form-panel">
      <div className="panel-header">
        <p className="section-kicker">Replay Layer</p>
        <h2>Filter snapshots</h2>
      </div>

      <div className="form-fields">
        <label className="form-label">
          Execution ID
          <input
            type="text"
            value={executionId}
            onChange={(e) => setExecutionId(e.target.value)}
            placeholder="Filter by execution ID..."
            className="form-input"
          />
        </label>
      </div>

      <button onClick={handleFilter} className="form-submit">
        {executionId ? "Filter" : "Show All"}
      </button>
    </div>
  );
}
