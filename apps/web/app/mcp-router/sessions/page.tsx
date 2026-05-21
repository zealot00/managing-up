"use client";

import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { listMCPSessions, type MCPSession } from "../../lib/api";
import { PageHeader } from "../../components/layout/PageHeader";
import { EmptyState } from "../../components/layout/EmptyState";
import Breadcrumb from "../../../components/Breadcrumb";
import { History } from "lucide-react";
import { useState } from "react";

export default function MCPSessionsPage() {
  const t = useTranslations("mcpRouter");
  const [agentFilter, setAgentFilter] = useState("");
  const { data, isLoading, isError } = useQuery({
    queryKey: ["mcp-sessions", agentFilter],
    queryFn: () => listMCPSessions(agentFilter || undefined, 100),
    refetchInterval: 30_000,
  });

  const sessions = data?.items ?? [];

  if (isLoading) {
    return (
      <>
        <Breadcrumb />
        <PageHeader
          eyebrow={t("eyebrow")}
          title={t("sessionsTitle")}
          description={t("sessionsDescription")}
        />
        <div className="skeleton-card" />
      </>
    );
  }

  if (isError) {
    return (
      <>
        <Breadcrumb />
        <PageHeader
          eyebrow={t("eyebrow")}
          title={t("sessionsTitle")}
          description={t("sessionsDescription")}
        />
        <div className="panel" role="alert">
          <p className="form-error">{t("noSessions")}</p>
        </div>
      </>
    );
  }

  return (
    <>
      <Breadcrumb />
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("sessionsTitle")}
        description={t("sessionsDescription")}
      />

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">{t("filter")}</p>
        </div>
        <form onSubmit={(e) => e.preventDefault()} className="form-fields">
          <label className="form-label">
            {t("agentId")}
            <input
              className="form-input"
              type="text"
              value={agentFilter}
              onChange={(e) => setAgentFilter(e.target.value)}
              placeholder={t("agentIdPlaceholder")}
            />
          </label>
        </form>
      </div>

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">{t("sessions")}</p>
          <h2 className="panel-title">{t("routingHistory")}</h2>
        </div>
        {isLoading ? (
          <div className="skeleton-card" />
        ) : sessions.length > 0 ? (
          <div className="table-wrapper">
            <table className="table">
              <thead>
                <tr>
<th scope="col">{t("sessionId")}</th>
                    <th scope="col">{t("agentId")}</th>
                    <th scope="col">{t("correlationId")}</th>
                    <th scope="col">{t("taskTypeCol")}</th>
                    <th scope="col">{t("policyDecision")}</th>
                    <th scope="col">{t("createdAt")}</th>
                </tr>
              </thead>
              <tbody>
                {sessions.map((session: MCPSession) => (
                  <tr key={session.id}>
                    <td className="cell-mono">
                      {session.id.slice(0, 8)}...
                    </td>
                    <td className="cell-mono">
                      {session.agent_id || "—"}
                    </td>
                    <td className="cell-mono">
                      {session.correlation_id || "—"}
                    </td>
                    <td>{session.task_type || "—"}</td>
                    <td>
                      <PolicyBadge decision={session.policy_decision} t={t} />
                    </td>
                    <td className="text-muted">
                      {new Date(session.created_at).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <EmptyState icon={<History size={32} />} title={t("noSessions")} description={t("noSessionsDesc")} />
        )}
      </div>
    </>
  );
}

function PolicyBadge({ decision, t }: { decision: string; t: ReturnType<typeof useTranslations<"mcpRouter">> }) {
  if (!decision) return <span className="text-muted">—</span>;
  try {
    const parsed = typeof decision === "string" ? JSON.parse(decision) : decision;
    const allowed = parsed.allowed;
    return (
      <span className={`badge ${allowed ? "badge-approved" : "badge-rejected"}`}>
        {allowed ? t("allowed") : t("denied")}
      </span>
    );
  } catch {
    return <span className="badge badge-muted">{String(decision)}</span>;
  }
}
