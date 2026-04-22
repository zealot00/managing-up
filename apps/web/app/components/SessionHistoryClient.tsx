"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { listGatewaySessions, GatewaySession } from "../lib/gateway-api";
import { PageHeader } from "./layout/PageHeader";
import { ListSkeleton } from "./layout/Skeleton";

const PAGE_SIZE = 20;

function SessionCard({ session }: { session: GatewaySession }) {
  const riskColors: Record<string, string> = {
    low: "bg-green-100 text-green-800",
    medium: "bg-yellow-100 text-yellow-800",
    high: "bg-red-100 text-red-800",
    critical: "bg-red-600 text-white",
  };

  const statusColors: Record<string, string> = {
    active: "bg-blue-100 text-blue-800",
    completed: "bg-gray-100 text-gray-800",
    cancelled: "bg-red-100 text-red-800",
  };

  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-shadow">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <span className="font-mono text-sm text-gray-600">{session.id.slice(0, 8)}</span>
          <span className={`px-2 py-0.5 rounded text-xs font-medium ${riskColors[session.risk_level] || "bg-gray-100 text-gray-800"}`}>
            {session.risk_level}
          </span>
          <span className={`px-2 py-0.5 rounded text-xs font-medium ${statusColors[session.status] || "bg-gray-100 text-gray-800"}`}>
            {session.status}
          </span>
        </div>
        <span className="text-sm text-gray-500">
          {new Date(session.started_at).toLocaleString()}
        </span>
      </div>
      <div className="grid grid-cols-2 gap-2 text-sm">
        <div>
          <span className="text-gray-500">Agent:</span>{" "}
          <span className="font-mono">{session.agent_id}</span>
        </div>
        <div>
          <span className="text-gray-500">Correlation:</span>{" "}
          <span className="font-mono">{session.correlation_id.slice(0, 8)}...</span>
        </div>
        <div>
          <span className="text-gray-500">Type:</span> <span>{session.session_type}</span>
        </div>
        <div>
          <span className="text-gray-500">Task:</span>{" "}
          <span>
            {(session.task_intent as { task_type?: string })?.task_type || "N/A"}
          </span>
        </div>
      </div>
      {session.policy_decision && (
        <div className="mt-2 pt-2 border-t border-gray-100">
          <span className="text-sm text-gray-500">Policy: </span>
          <span className="text-sm">
            {(session.policy_decision as { allowed?: boolean; reasons?: string[] })
              ?.allowed === false
              ? "DENIED"
              : "ALLOWED"}
          </span>
        </div>
      )}
    </div>
  );
}

export default function SessionHistoryClient() {
  const [agentFilter, setAgentFilter] = useState("");
  const [displayCount, setDisplayCount] = useState(PAGE_SIZE);

  const { data: sessionsData, isLoading } = useQuery({
    queryKey: ["gateway-sessions", agentFilter],
    queryFn: () => listGatewaySessions({ agent_id: agentFilter || undefined, limit: 100 }),
  });

  const sessions = Array.isArray(sessionsData) ? sessionsData : sessionsData?.items ?? [];

  const displayedSessions = sessions.slice(0, displayCount);
  const hasMore = displayCount < sessions.length;

  return (
    <div className="space-y-4">
      <PageHeader
        title="Gateway Sessions"
        description="View gateway session history and policy decisions"
      />

      <div className="flex items-center gap-4">
        <input
          type="text"
          placeholder="Filter by agent ID..."
          value={agentFilter}
          onChange={(e) => setAgentFilter(e.target.value)}
          className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      {isLoading ? (
        <ListSkeleton rows={5} />
      ) : sessions.length === 0 ? (
        <div className="text-center py-12 text-gray-500">
          No sessions found
        </div>
      ) : (
        <>
          <div className="space-y-3">
            {displayedSessions.map((session) => (
              <SessionCard key={session.id} session={session} />
            ))}
          </div>
          {hasMore && (
            <div className="flex justify-center">
              <button
                onClick={() => setDisplayCount((c) => c + PAGE_SIZE)}
                className="px-4 py-2 text-sm text-blue-600 hover:text-blue-800"
              >
                Load more
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}