"use client";

import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { getMCPRouterCatalog, matchMCPRouter, type MCPRouterCatalogEntry } from "../lib/api";
import { PageHeader } from "../components/layout/PageHeader";
import { EmptyState } from "../components/layout/EmptyState";
import Breadcrumb from "../../components/Breadcrumb";
import { useState } from "react";
import { Server } from "lucide-react";

export function MCPRouterDashboardClient() {
  const t = useTranslations("mcpRouter");
  const { data: catalog, isLoading, isError } = useQuery({
    queryKey: ["mcp-router-catalog"],
    queryFn: getMCPRouterCatalog,
  });

  const stats = {
    total: catalog?.length ?? 0,
    active: catalog?.filter(s => s.status === "active").length ?? 0,
    avgTrust: catalog?.length ? (catalog.reduce((sum, s) => sum + s.trust_score, 0) / catalog.length) : 0,
    totalUse: catalog?.reduce((sum, s) => sum + s.use_count, 0) ?? 0,
  };

  if (isLoading) {
    return (
      <>
        <Breadcrumb />
        <PageHeader
          eyebrow={t("eyebrow")}
          title={t("title")}
          description={t("description")}
        />
        <div className="dashboard-stats">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="dashboard-stat-card">
              <div className="loading-pulse loading-pulse-medium" style={{ width: 60, height: 16, marginBottom: 8 }} />
              <div className="loading-pulse" style={{ width: 40, height: 28 }} />
            </div>
          ))}
        </div>
        <div className="skeleton-grid">
          {[1, 2].map((i) => (
            <div key={i} className="skeleton-card" />
          ))}
        </div>
      </>
    );
  }

  if (isError) {
    return (
      <>
        <Breadcrumb />
        <PageHeader
          eyebrow={t("eyebrow")}
          title={t("title")}
          description={t("description")}
        />
        <div className="panel" role="alert">
          <p className="form-error">{t("noServers")}</p>
        </div>
      </>
    );
  }

  return (
    <>
      <Breadcrumb />
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={t("description")}
      />

      <div className="dashboard-stats">
        <article className="dashboard-stat-card">
          <div className="dashboard-stat-value">{stats.total}</div>
          <div className="dashboard-stat-label">{t("totalServers")}</div>
        </article>
        <article className="dashboard-stat-card">
          <div className="dashboard-stat-value">{stats.active}</div>
          <div className="dashboard-stat-label">{t("active")}</div>
        </article>
        <article className="dashboard-stat-card">
          <div className="dashboard-stat-value">{stats.avgTrust.toFixed(2)}</div>
          <div className="dashboard-stat-label">{t("avgTrustScore")}</div>
        </article>
        <article className="dashboard-stat-card">
          <div className="dashboard-stat-value">{stats.totalUse}</div>
          <div className="dashboard-stat-label">{t("totalInvocations")}</div>
        </article>
      </div>

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">{t("routing")}</p>
          <h2 className="panel-title">{t("routeTest")}</h2>
        </div>
        <RouteTestPanel t={t} />
      </div>

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">{t("catalog")}</p>
          <h2 className="panel-title">{t("routerCatalog")}</h2>
        </div>
        {isLoading ? (
          <div className="skeleton-card" />
        ) : catalog && catalog.length > 0 ? (
          <div className="table-wrapper">
            <table className="table">
              <thead>
                <tr>
<th scope="col">{t("name")}</th>
                    <th scope="col">{t("transport")}</th>
                    <th scope="col">{t("taskTypes")}</th>
                    <th scope="col">{t("trust")}</th>
                    <th scope="col">{t("invocations")}</th>
                    <th scope="col">{t("status")}</th>
                </tr>
              </thead>
              <tbody>
                {catalog.map((server) => (
                  <tr key={server.id}>
                    <td>
                      <strong>{server.name}</strong>
                      {server.description && (
                        <><br /><span className="text-muted">{server.description}</span></>
                      )}
                    </td>
                    <td><span className="badge badge-muted">{server.transport_type}</span></td>
                    <td>
                      {server.task_types?.length ? (
                        server.task_types.map((type) => (
                          <span key={type} className="badge badge-draft">{type}</span>
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
          <EmptyState icon={<Server size={32} />} title={t("noServers")} description={t("noServersDesc")} />
        )}
      </div>
    </>
  );
}

function RouteTestPanel({ t }: { t: ReturnType<typeof useTranslations<"mcpRouter">> }) {
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
          {t("taskType")}
          <input className="form-input" type="text" value={taskType} onChange={(e) => setTaskType(e.target.value)} placeholder={t("taskTypePlaceholder")} />
        </label>
        <label className="form-label">
          {t("tags")}
          <input className="form-input" type="text" value={tagsInput} onChange={(e) => setTagsInput(e.target.value)} placeholder={t("tagsPlaceholder")} />
        </label>
      </div>
      <button type="submit" className="btn btn-primary btn-sm" disabled={!taskType || testing}>
        {testing ? t("testing") : t("testRoute")}
      </button>
      {result && (
        <div className="notice-box">
          {result.matched ? (
            <p>{t("matched")}: <strong>{result.server_name}</strong> (score: {result.score?.toFixed(2)})</p>
          ) : (
            <p className="text-muted">{t("noMatch")}</p>
          )}
        </div>
      )}
    </form>
  );
}
