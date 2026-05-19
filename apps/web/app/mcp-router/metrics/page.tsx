"use client";

import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { getMCPRouterCatalog } from "../../lib/api";
import { PageHeader } from "../../components/layout/PageHeader";
import { EmptyState } from "../../components/layout/EmptyState";
import Breadcrumb from "../../../components/Breadcrumb";
import { BarChart3 } from "lucide-react";

export default function MCPRouterMetricsPage() {
  const t = useTranslations("mcpRouter");
  const { data: catalog, isLoading, isError } = useQuery({
    queryKey: ["mcp-router-catalog"],
    queryFn: getMCPRouterCatalog,
  });

  const sortedByUseCount = [...(catalog ?? [])].sort((a, b) => b.use_count - a.use_count);

  if (isLoading) {
    return (
      <>
        <Breadcrumb />
        <PageHeader
          eyebrow={t("eyebrow")}
          title={t("metricsTitle")}
          description={t("metricsDescription")}
        />
        <div className="skeleton-grid">
          {[1, 2, 3].map((i) => (
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
          title={t("metricsTitle")}
          description={t("metricsDescription")}
        />
        <div className="panel" role="alert">
          <p className="form-error">{t("noServerData")}</p>
        </div>
      </>
    );
  }

  return (
    <>
      <Breadcrumb />
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("metricsTitle")}
        description={t("metricsDescription")}
      />

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">{t("usage")}</p>
          <h2 className="panel-title">{t("serverUsageRanking")}</h2>
        </div>
        {isLoading ? (
          <div className="skeleton-card" />
        ) : sortedByUseCount.length > 0 ? (
          <div className="table-wrapper">
            <table className="table">
              <thead>
                <tr>
<th scope="col">{t("rank")}</th>
                    <th scope="col">{t("server")}</th>
                    <th scope="col">{t("transport")}</th>
                    <th scope="col">{t("trustScore")}</th>
                    <th scope="col">{t("invocations")}</th>
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
          <EmptyState icon={<BarChart3 size={32} />} title={t("noServerData")} description={t("noServerDataDesc")} />
        )}
      </div>

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">{t("distribution")}</p>
          <h2 className="panel-title">{t("trustScoreDistribution")}</h2>
        </div>
        {catalog && catalog.length > 0 ? (
          <div className="table-wrapper">
            <table className="table">
              <thead>
                <tr>
                  <th>{t("server")}</th>
                  <th>{t("trustScore")}</th>
                  <th>{t("bar")}</th>
                </tr>
              </thead>
              <tbody>
                {catalog.map((server) => (
                  <tr key={server.id}>
                    <td>{server.name}</td>
                    <td>{server.trust_score.toFixed(2)}</td>
                    <td style={{ width: "40%" }}>
                      <div className="progress-bar">
                        <div
                          className="progress-bar-fill"
                          style={{
                            width: `${Math.max(server.trust_score * 100, 2)}%`,
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
          <EmptyState icon={<BarChart3 size={32} />} title={t("noData")} />
        )}
      </div>

      <div className="panel">
        <div className="panel-header">
          <p className="section-kicker">{t("observability")}</p>
          <h2 className="panel-title">{t("prometheusMetrics")}</h2>
        </div>
        <p className="text-muted" style={{ marginBottom: "var(--space-5)" }}>
          {t("prometheusDesc")}
        </p>
        <div className="table-wrapper">
          <table className="table">
            <thead>
              <tr>
                <th>{t("metric")}</th>
                <th>{t("metricName")}</th>
                <th>{t("metricDescription")}</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td><strong>{t("totalRequests")}</strong></td>
                <td><code>mcp_router_requests_total</code></td>
                <td className="text-muted">{t("totalRequestsDesc")}</td>
              </tr>
              <tr>
                <td><strong>{t("requestDuration")}</strong></td>
                <td><code>mcp_router_request_duration_seconds</code></td>
                <td className="text-muted">{t("requestDurationDesc")}</td>
              </tr>
              <tr>
                <td><strong>{t("matchFailures")}</strong></td>
                <td><code>mcp_router_match_failures_total</code></td>
                <td className="text-muted">{t("matchFailuresDesc")}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}
