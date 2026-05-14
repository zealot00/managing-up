"use client";

import { useQuery } from "@tanstack/react-query";
import { getMCPRouterCatalog } from "../../lib/api";
import { PageHeader } from "../../components/layout/PageHeader";
import { EmptyState } from "../../components/layout/EmptyState";

export default function MCPRouterMetricsPage() {
  const { data: catalog, isLoading } = useQuery({
    queryKey: ["mcp-router-catalog"],
    queryFn: getMCPRouterCatalog,
  });

  const sortedByUseCount = [...(catalog ?? [])].sort((a, b) => b.use_count - a.use_count);

  return (
    <>
      <PageHeader
        eyebrow="Operations"
        title="MCP Router Metrics"
        description="Usage statistics and trust score distribution for routed MCP servers."
      />

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">Usage</p>
          <h2 className="panel-title">Server Usage Ranking</h2>
        </div>
        {isLoading ? (
          <p className="empty-note">Loading...</p>
        ) : sortedByUseCount.length > 0 ? (
          <div className="table-wrapper">
            <table className="table">
              <thead>
                <tr>
                  <th>Rank</th>
                  <th>Server</th>
                  <th>Transport</th>
                  <th>Trust Score</th>
                  <th>Invocations</th>
                </tr>
              </thead>
              <tbody>
                {sortedByUseCount.map((server, idx) => (
                  <tr key={server.id}>
                    <td>#{idx + 1}</td>
                    <td><strong>{server.name}</strong></td>
                    <td><span className="badge badge-muted">{server.transport_type}</span></td>
                    <td>{server.trust_score.toFixed(2)}</td>
                    <td><strong>{server.use_count}</strong></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <EmptyState title="No server data" description="No MCP servers have been routed yet." />
        )}
      </div>

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">Distribution</p>
          <h2 className="panel-title">Trust Score Distribution</h2>
        </div>
        {catalog && catalog.length > 0 ? (
          <div className="table-wrapper">
            <table className="table">
              <thead>
                <tr>
                  <th>Server</th>
                  <th>Trust Score</th>
                  <th>Bar</th>
                </tr>
              </thead>
              <tbody>
                {catalog.map((server) => (
                  <tr key={server.id}>
                    <td>{server.name}</td>
                    <td>{server.trust_score.toFixed(2)}</td>
                    <td style={{ width: "40%" }}>
                      <div style={{ background: "var(--surface)", borderRadius: "var(--radius-full)", height: "0.5rem", overflow: "hidden" }}>
                        <div
                          style={{
                            height: "100%",
                            background: "var(--primary)",
                            borderRadius: "var(--radius-full)",
                            width: `${Math.max(server.trust_score * 100, 2)}%`,
                            transition: "width 0.3s ease",
                          }}
                        />
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <EmptyState title="No data" />
        )}
      </div>

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">Observability</p>
          <h2 className="panel-title">Prometheus Metrics</h2>
        </div>
        <p style={{ color: "var(--muted)", fontSize: "var(--text-sm)", marginBottom: "var(--space-5)" }}>
          Detailed metrics are exposed in Prometheus format at the <code>/metrics</code> endpoint.
        </p>
        <div className="table-wrapper">
          <table className="table">
            <thead>
              <tr>
                <th>Metric</th>
                <th>Name</th>
                <th>Description</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td><strong>Total Requests</strong></td>
                <td><code>mcp_router_requests_total</code></td>
                <td style={{ color: "var(--muted)" }}>Counter of all MCP router requests</td>
              </tr>
              <tr>
                <td><strong>Request Duration</strong></td>
                <td><code>mcp_router_request_duration_seconds</code></td>
                <td style={{ color: "var(--muted)" }}>Histogram of request latency</td>
              </tr>
              <tr>
                <td><strong>Match Failures</strong></td>
                <td><code>mcp_router_match_failures_total</code></td>
                <td style={{ color: "var(--muted)" }}>Counter of route match failures</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}
