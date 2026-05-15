"use client";

import { useQuery } from "@tanstack/react-query";
import { listMCPSessions, type MCPSession } from "../../lib/api";
import { PageHeader } from "../../components/layout/PageHeader";
import { EmptyState } from "../../components/layout/EmptyState";
import { useState } from "react";

export default function MCPSessionsPage() {
  const [agentFilter, setAgentFilter] = useState("");
  const { data, isLoading } = useQuery({
    queryKey: ["mcp-sessions", agentFilter],
    queryFn: () => listMCPSessions(agentFilter || undefined, 100),
  });

  const sessions = data?.items ?? [];

  return (
    <>
      <PageHeader
        eyebrow="Operations"
        title="MCP Router Sessions"
        description="History of routing sessions with policy decisions and match results."
      />

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">Filter</p>
        </div>
        <form onSubmit={(e) => e.preventDefault()} className="form-fields">
          <label className="form-label">
            Agent ID
            <input
              className="form-input"
              type="text"
              value={agentFilter}
              onChange={(e) => setAgentFilter(e.target.value)}
              placeholder="Enter agent ID to filter"
            />
          </label>
        </form>
      </div>

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">Sessions</p>
          <h2 className="panel-title">Routing History</h2>
        </div>
        {isLoading ? (
          <p className="empty-note">Loading...</p>
        ) : sessions.length > 0 ? (
          <div className="table-wrapper">
            <table className="table">
              <thead>
                <tr>
                  <th>Session ID</th>
                  <th>Agent ID</th>
                  <th>Correlation ID</th>
                  <th>Task Type</th>
                  <th>Policy Decision</th>
                  <th>Created At</th>
                </tr>
              </thead>
              <tbody>
                {sessions.map((session: MCPSession) => (
                  <tr key={session.id}>
                    <td style={{ fontFamily: "var(--font-mono, monospace)", fontSize: "var(--text-xs)" }}>
                      {session.id.slice(0, 8)}...
                    </td>
                    <td style={{ fontFamily: "var(--font-mono, monospace)", fontSize: "var(--text-xs)" }}>
                      {session.agent_id || "—"}
                    </td>
                    <td style={{ fontFamily: "var(--font-mono, monospace)", fontSize: "var(--text-xs)" }}>
                      {session.correlation_id || "—"}
                    </td>
                    <td>{session.task_type || "—"}</td>
                    <td>
                      <PolicyBadge decision={session.policy_decision} />
                    </td>
                    <td style={{ color: "var(--muted)", fontSize: "var(--text-sm)" }}>
                      {new Date(session.created_at).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <EmptyState title="No sessions found" description="Routing sessions will appear here when MCP tools are invoked." />
        )}
      </div>
    </>
  );
}

function PolicyBadge({ decision }: { decision: string }) {
  if (!decision) return <span style={{ color: "var(--muted)" }}>—</span>;
  try {
    const parsed = typeof decision === "string" ? JSON.parse(decision) : decision;
    const allowed = parsed.allowed;
    return (
      <span className={`badge ${allowed ? "badge-approved" : "badge-rejected"}`}>
        {allowed ? "Allowed" : "Denied"}
      </span>
    );
  } catch {
    return <span className="badge badge-muted">{String(decision)}</span>;
  }
}
