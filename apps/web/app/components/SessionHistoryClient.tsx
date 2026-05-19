"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { listGatewaySessions, GatewaySession } from "../lib/gateway-api";
import { PageHeader } from "./layout/PageHeader";
import { EmptyState } from "./layout/EmptyState";
import { ListSkeleton } from "./layout/Skeleton";
import { Badge } from "./ui/Badge";
import { Search } from "lucide-react";

const PAGE_SIZE = 20;

const riskBadgeVariant: Record<string, "low" | "high" | "pending" | "draft"> = {
  low: "low",
  medium: "pending",
  high: "high",
  critical: "high",
};

const statusBadgeVariant: Record<string, "active" | "completed" | "failed" | "draft"> = {
  active: "active",
  completed: "completed",
  cancelled: "failed",
};

function SessionCard({ session, t }: { session: GatewaySession; t: ReturnType<typeof useTranslations<"sessionHistory">> }) {
  return (
    <div className="list-card" style={{ flexDirection: "column", alignItems: "stretch" }}>
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "var(--space-2)" }}>
        <div style={{ display: "flex", alignItems: "center", gap: "var(--space-2)" }}>
          <span className="cell-mono">
            {session.id.slice(0, 8)}
          </span>
          <Badge variant={riskBadgeVariant[session.risk_level] || "draft"}>
            {session.risk_level}
          </Badge>
          <Badge variant={statusBadgeVariant[session.status] || "draft"}>
            {session.status}
          </Badge>
        </div>
        <span className="text-muted">
          {new Date(session.started_at).toLocaleString()}
        </span>
      </div>
      <div style={{ display: "grid", gridTemplateColumns: "repeat(2, 1fr)", gap: "var(--space-2)", fontSize: "var(--text-sm)" }}>
        <div>
          <span className="text-muted">{t("agent")}:</span>{" "}
          <span className="font-mono">{session.agent_id}</span>
        </div>
        <div>
          <span className="text-muted">{t("correlation")}:</span>{" "}
          <span className="font-mono">{session.correlation_id.slice(0, 8)}...</span>
        </div>
        <div>
          <span className="text-muted">{t("type")}:</span> <span>{session.session_type}</span>
        </div>
        <div>
          <span className="text-muted">{t("task")}:</span>{" "}
          <span>{(session.task_intent as { task_type?: string })?.task_type || t("na")}</span>
        </div>
      </div>
      {session.policy_decision && (
        <div style={{ marginTop: "var(--space-2)", paddingTop: "var(--space-2)", borderTop: "1px solid var(--line)", fontSize: "var(--text-sm)" }}>
          <span className="text-muted">{t("policy")}: </span>
          <span style={{ fontWeight: 600, color: (session.policy_decision as { allowed?: boolean })?.allowed === false ? "var(--danger)" : "var(--success)" }}>
            {(session.policy_decision as { allowed?: boolean; reasons?: string[] })
              ?.allowed === false
              ? t("denied")
              : t("allowed")}
          </span>
        </div>
      )}
    </div>
  );
}

export default function SessionHistoryClient() {
  const t = useTranslations("sessionHistory");
  const [agentFilter, setAgentFilter] = useState("");
  const [displayCount, setDisplayCount] = useState(PAGE_SIZE);

  const { data: sessionsData, isLoading, isError } = useQuery({
    queryKey: ["gateway-sessions", agentFilter],
    queryFn: () => listGatewaySessions({ agent_id: agentFilter || undefined, limit: 100 }),
  });

  const sessions = Array.isArray(sessionsData) ? sessionsData : sessionsData?.items ?? [];

  const displayedSessions = sessions.slice(0, displayCount);
  const hasMore = displayCount < sessions.length;

  return (
    <>
      <PageHeader
        title={t("title")}
        description={t("description")}
      />

      <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
        <div className="form-group" style={{ flex: 1, maxWidth: 400 }}>
          <div style={{ position: "relative" }}>
            <Search size={16} className="search-icon" />
            <input
              type="text"
              placeholder={t("filterByAgentId")}
              value={agentFilter}
              onChange={(e) => setAgentFilter(e.target.value)}
              className="search-input"
            />
          </div>
        </div>
      </div>

      {isLoading ? (
        <ListSkeleton rows={5} />
      ) : isError ? (
        <div className="panel" role="alert">
          <p className="form-error">{t("noSessions")}</p>
        </div>
      ) : sessions.length === 0 ? (
        <EmptyState
          icon={<Search size={32} />}
          title={t("noSessions")}
          description={t("noSessionsDesc")}
        />
      ) : (
        <>
          <div className="list" style={{ marginTop: "var(--space-5)" }}>
            {displayedSessions.map((session) => (
              <SessionCard key={session.id} session={session} t={t} />
            ))}
          </div>
          {hasMore && (
            <div style={{ display: "flex", justifyContent: "center", marginTop: "var(--space-5)" }}>
              <button
                onClick={() => setDisplayCount((c) => c + PAGE_SIZE)}
                className="btn btn-ghost btn-sm"
              >
                {t("loadMore")}
              </button>
            </div>
          )}
        </>
      )}
    </>
  );
}
