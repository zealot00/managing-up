"use client";

import { useQuery } from "@tanstack/react-query";
import { getMCPRouterCatalog, matchMCPRouter, type MCPRouterCatalogEntry } from "../lib/api";
import { PageHeader } from "../components/layout/PageHeader";
import { EmptyState } from "../components/layout/EmptyState";
import { useState } from "react";

export function MCPRouterDashboardClient() {
  const { data: catalog, isLoading } = useQuery({
    queryKey: ["mcp-router-catalog"],
    queryFn: getMCPRouterCatalog,
  });

  const stats = {
    total: catalog?.length ?? 0,
    active: catalog?.filter(s => s.status === "active").length ?? 0,
    avgTrust: catalog?.length ? (catalog.reduce((sum, s) => sum + s.trust_score, 0) / catalog.length) : 0,
    totalUse: catalog?.reduce((sum, s) => sum + s.use_count, 0) ?? 0,
  };

  return (
    <>
      <PageHeader
        eyebrow="Operations"
        title="MCP Router"
        description="Route agent requests to the best MCP server based on task type, tags, and trust score."
      />

      <div className="dashboard-stats">
        <article className="dashboard-stat-card">
          <div className="dashboard-stat-value">{stats.total}</div>
          <div className="dashboard-stat-label">Total Servers</div>
        </article>
        <article className="dashboard-stat-card">
          <div className="dashboard-stat-value">{stats.active}</div>
          <div className="dashboard-stat-label">Active</div>
        </article>
        <article className="dashboard-stat-card">
          <div className="dashboard-stat-value">{stats.avgTrust.toFixed(2)}</div>
          <div className="dashboard-stat-label">Avg Trust Score</div>
        </article>
        <article className="dashboard-stat-card">
          <div className="dashboard-stat-value">{stats.totalUse}</div>
          <div className="dashboard-stat-label">Total Invocations</div>
        </article>
      </div>

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">Routing</p>
          <h2 className="panel-title">Route Test</h2>
        </div>
        <RouteTestPanel />
      </div>

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">Catalog</p>
          <h2 className="panel-title">Router Catalog</h2>
        </div>
        {isLoading ? (
          <p className="empty-note">Loading...</p>
        ) : catalog && catalog.length > 0 ? (
          <div className="table-wrapper">
            <table className="table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Transport</th>
                  <th>Task Types</th>
                  <th>Trust</th>
                  <th>Invocations</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {catalog.map((server) => (
                  <tr key={server.id}>
                    <td>
                      <strong>{server.name}</strong>
                      {server.description && (
                        <><br /><span style={{ color: "var(--muted)", fontSize: "var(--text-xs)" }}>{server.description}</span></>
                      )}
                    </td>
                    <td><span className="badge badge-muted">{server.transport_type}</span></td>
                    <td>
                      {server.task_types?.length ? (
                        server.task_types.map((type) => (
                          <span key={type} className="badge badge-draft" style={{ marginRight: "var(--space-1)" }}>{type}</span>
                        ))
                      ) : "—"}
                    </td>
                    <td>{server.trust_score.toFixed(2)}</td>
                    <td>{server.use_count}</td>
                    <td>
                      <span className={`badge ${server.status === "active" ? "badge-active" : "badge-pending"}`}>
                        {server.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <EmptyState title="No servers in catalog" description="Approve MCP servers to add them to the router." />
        )}
      </div>
    </>
  );
}

function RouteTestPanel() {
  const [taskType, setTaskType] = useState("");
  const [tagsInput, setTagsInput] = useState("");
  const [result, setResult] = useState<{ matched: boolean; server_name?: string; score?: number } | null>(null);
  const [testing, setTesting] = useState(false);

  async function handleTest(e: React.FormEvent) {
    e.preventDefault();
    if (!taskType) return;
    setTesting(true);
    setResult(null);
    try {
      const tags = tagsInput ? tagsInput.split(",").map(t => t.trim()).filter(Boolean) : undefined;
      const res = await matchMCPRouter(taskType, tags);
      if (res) {
        setResult({ matched: res.matched, server_name: res.target?.server_name, score: res.match_score });
      } else {
        setResult({ matched: false });
      }
    } catch {
      setResult({ matched: false });
    } finally {
      setTesting(false);
    }
  }

  return (
    <form onSubmit={handleTest}>
      <div className="form-fields">
        <label className="form-label">
          Task Type
          <input className="form-input" type="text" value={taskType} onChange={(e) => setTaskType(e.target.value)} placeholder="e.g. code_generation" />
        </label>
        <label className="form-label">
          Tags (comma-separated)
          <input className="form-input" type="text" value={tagsInput} onChange={(e) => setTagsInput(e.target.value)} placeholder="e.g. python,backend" />
        </label>
      </div>
      <button type="submit" className="form-submit" disabled={!taskType || testing} style={{ marginTop: "var(--space-4)" }}>
        {testing ? "Testing..." : "Test Route"}
      </button>
      {result && (
        <div style={{ marginTop: "var(--space-4)", padding: "var(--space-4)", border: "1px solid var(--line)", borderRadius: "var(--radius-md)" }}>
          {result.matched ? (
            <p>Matched: <strong>{result.server_name}</strong> (score: {result.score?.toFixed(2)})</p>
          ) : (
            <p style={{ color: "var(--muted)" }}>No matching server found.</p>
          )}
        </div>
      )}
    </form>
  );
}
