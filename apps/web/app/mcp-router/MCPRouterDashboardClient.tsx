"use client";

import { useQuery } from "@tanstack/react-query";
import { getMCPRouterCatalog } from "../lib/api";

export function MCPRouterDashboardClient() {
  const { data: catalog, isLoading } = useQuery({
    queryKey: ["mcp-router-catalog"],
    queryFn: getMCPRouterCatalog,
  });

  if (isLoading) {
    return <div className="p-6">Loading...</div>;
  }

  const stats = {
    total: catalog?.length ?? 0,
    active: catalog?.filter(s => s.status === "active").length ?? 0,
    avgTrust: (catalog?.reduce((sum, s) => sum + s.trust_score, 0) ?? 0) / (catalog?.length ?? 1),
  };

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">MCP Router Dashboard</h1>
      <div className="grid grid-cols-3 gap-4 mb-8">
        <StatCard label="Total Servers" value={stats.total} />
        <StatCard label="Active" value={stats.active} />
        <StatCard label="Avg Trust Score" value={stats.avgTrust.toFixed(2)} />
      </div>
      
      <h2 className="text-lg font-semibold mb-4">Router Catalog</h2>
      <div className="space-y-2">
        {catalog?.map((server) => (
          <div key={server.id} className="card p-4">
            <div className="flex justify-between items-start">
              <div>
                <h3 className="font-medium">{server.name}</h3>
                <p className="text-sm text-muted">{server.description}</p>
              </div>
              <div className="flex gap-2">
                <span className="badge">{server.transport_type}</span>
                <span className="badge">{server.status}</span>
              </div>
            </div>
            <div className="mt-2 flex gap-2">
              {server.task_types?.map((type) => (
                <span key={type} className="badge badge-outline">{type}</span>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="card p-4">
      <div className="text-sm text-muted">{label}</div>
      <div className="text-2xl font-bold">{value}</div>
    </div>
  );
}
